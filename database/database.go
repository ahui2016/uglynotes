package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/stmt"
	"github.com/ahui2016/uglynotes/stringset"
	"github.com/ahui2016/uglynotes/util"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/ianbruene/go-difflib/difflib"
	_ "github.com/mattn/go-sqlite3"
)

const cookieName = "uglynotesCookie"

var config = settings.Config

type (
	Note       = model.Note
	NoteType   = model.NoteType
	History    = model.History
	Tag        = model.Tag
	TagGroup   = model.TagGroup
	IncreaseID = model.IncreaseID
	Set        = stringset.Set
	Stmt       = sql.Stmt
)

var stmtGetTagNamesByNote, stmtGetNote, stmtGetNotes, stmtGetDeletedNotes,
	stmtGetPatchesByNote *Stmt

type TX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*Stmt, error)
}

type Row interface {
	Scan(...interface{}) error
}

type DB2 struct {
	path string
	DB   *sql.DB
	Sess *session.Store
	sync.Mutex
}

func (db *DB2) Open(dbPath string) (err error) {
	if db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1"); err != nil {
		return err
	}
	if _, err = db.DB.Exec(stmt.CreateTables); err != nil {
		return err
	}
	db.path = dbPath
	// db.Sess = session.New(session.Config{
	// 	Expiration: mustParseDuration(config.MaxAge),
	// 	CookieName: cookieName,
	// })
	db.prepareStatements()
	err1 := initFirstID(db.DB)
	err2 := initTotalSize(db.DB)
	return util.WrapErrors(err1, err2)
}
func (db *DB2) Close() error {
	closeStatements()
	return db.DB.Close()
}

func (db *DB2) mustBegin() *sql.Tx {
	tx, err := db.DB.Begin()
	util.Panic(err)
	return tx
}

func mustPrepare(tx TX, query string) *Stmt {
	stmt, err := tx.Prepare(query)
	util.Panic(err)
	return stmt
}

func (db *DB2) prepareStatements() {
	stmtGetTagNamesByNote = mustPrepare(db.DB, stmt.GetTagNamesByNote)
	stmtGetNote = mustPrepare(db.DB, stmt.GetNote)
	stmtGetNotes = mustPrepare(db.DB, stmt.GetNotes)
	stmtGetDeletedNotes = mustPrepare(db.DB, stmt.GetDeletedNotes)
	stmtGetPatchesByNote = mustPrepare(db.DB, stmt.GetPatchesByNote)
}

func closeStatements() {
	stmtGetTagNamesByNote.Close()
	stmtGetNote.Close()
	stmtGetNotes.Close()
	stmtGetDeletedNotes.Close()
	stmtGetPatchesByNote.Close()
}

func (db *DB2) ImportNotes(notes []Note) (err error) {
	tx := db.mustBegin()
	defer tx.Rollback()

	stmtInsertNote := mustPrepare(tx, stmt.InsertNote)
	defer stmtInsertNote.Close()

	stmtInsertPatch := mustPrepare(tx, stmt.InsertPatch)
	defer stmtInsertPatch.Close()
	stmtInsertNotePatch := mustPrepare(tx, stmt.InsertNotePatch)
	defer stmtInsertNotePatch.Close()

	stmtGetTagID := mustPrepare(tx, stmt.GetTagID)
	defer stmtGetTagID.Close()
	stmtInsertTag := mustPrepare(tx, stmt.InsertTag)
	defer stmtInsertTag.Close()
	stmtInsertNoteTag := mustPrepare(tx, stmt.InsertNoteTag)
	defer stmtInsertNoteTag.Close()

	stmtGetTagGroupID := mustPrepare(tx, stmt.GetTagGroupID)
	defer stmtGetTagID.Close()
	stmtInsertTagGroup := mustPrepare(tx, stmt.InsertTagGroup)
	defer stmtInsertTagGroup.Close()
	stmtUpdateTagGroupNow := mustPrepare(tx, stmt.UpdateTagGroupNow)
	defer stmtUpdateTagGroupNow.Close()

	for _, note := range notes {
		if err = importNote(stmtInsertNote, note); err != nil {
			return fmt.Errorf("importNote: %v", err)
		}
		if err = importTagGroup(stmtGetTagGroupID, stmtInsertTagGroup,
			stmtUpdateTagGroupNow, note.Tags); err != nil {
			return fmt.Errorf("importTagGroup: %v", err)
		}
		if err = importTags(stmtGetTagID, stmtInsertTag, stmtInsertNoteTag,
			note.ID, note.Tags); err != nil {
			return fmt.Errorf("importTags: %v", err)
		}
		if err = importPatches(stmtInsertPatch, stmtInsertNotePatch,
			note.ID, note.Patches); err != nil {
			return fmt.Errorf("importPatches: %v", err)
		}
		if err = increaseTotalSize(tx, note.Size); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func importNote(stmtAdd *Stmt, note Note) (err error) {
	_, err = stmtAdd.Exec(
		note.ID,
		note.Type,
		note.Title,
		note.Size,
		btoi(note.Deleted),
		"",
		note.CreatedAt,
		note.UpdatedAt,
	)
	return
}

func importTagGroup(
	stmtGet, stmtAdd, stmtUpdate *Stmt, tags []string) error {
	groupID, err := getText1(stmtGet, util.MustMarshal(tags))
	if err == sql.ErrNoRows {
		g := model.NewTagGroup(tags)
		_, err = stmtAdd.Exec(g.ID, util.MustMarshal(g.Tags),
			btoi(g.Protected), g.CreatedAt, g.UpdatedAt)
		return err
	}
	return updateNow(stmtUpdate, groupID)
}

func updateNow(stmtUpdate *Stmt, id string) error {
	_, err := stmtUpdate.Exec(model.TimeNow(), id)
	return err
}

// getText1 gets a text value by one argument.
func getText1(stmtGet *Stmt, arg interface{}) (text string, err error) {
	row := stmtGet.QueryRow(arg)
	err = row.Scan(&text)
	return
}

func importTags(stmtGet, stmtAdd, stmt3 *Stmt, noteID string,
	tags []string) (err error) {
	for _, tagName := range tags {
		err = importTag(stmtGet, stmtAdd, stmt3, noteID, tagName)
		if err != nil {
			return
		}
	}
	return
}

func importTag(
	stmtGet, stmtAdd, stmt3 *Stmt, noteID string, name string) error {
	tagID, err := getText1(stmtGet, name)
	if err == sql.ErrNoRows {
		tagID = model.RandomID()
		_, err = stmtAdd.Exec(tagID, name, model.TimeNow())
	}
	if err != nil {
		return err
	}
	_, err = stmt3.Exec(noteID, tagID)
	return err
}

func importPatches(stmt1, stmt2 *Stmt, noteID string, patches []string) (
	err error) {
	for _, diff := range patches {
		if err = insertPatch(stmt1, stmt2, noteID, diff); err != nil {
			return
		}
	}
	return
}

func insertPatch(stmt1, stmt2 *Stmt, noteID, diff string) error {
	patchID := model.NextTimeID()
	_, err1 := stmt1.Exec(patchID, diff)
	_, err2 := stmt2.Exec(noteID, patchID)
	return util.WrapErrors(err1, err2)
}

func (db *DB2) FillGroups(groups []TagGroup) error {
	questions := make([]string, 0, len(groups))
	values := make([]interface{}, 0, len(groups)*5)
	for _, group := range groups {
		questions = append(questions, "(?,?,?,?,?)")
		values = append(values, group.ID)
		values = append(values, util.MustMarshal(group.Tags))
		values = append(values, btoi(group.Protected))
		values = append(values, group.CreatedAt)
		values = append(values, group.UpdatedAt)
	}
	stmt := fmt.Sprintf(
		"INSERT INTO taggroup (id, tags, protected, created_at, updated_at) VALUES %s",
		strings.Join(questions, ","))
	_, err := db.DB.Exec(stmt, values...)
	return err
}
func (db *DB2) DropTagGroup() error {
	_, err := db.DB.Exec("DROP TABLE IF EXISTS taggroup")
	return err
}

func (db *DB2) AllTagGroups() (groups []TagGroup, err error) {
	rows, err := db.DB.Query("SELECT * FROM taggroup")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var group *TagGroup
		group, err = scanTagGroup(rows)
		if err != nil {
			return
		}
		groups = append(groups, *group)
	}
	err = rows.Err()
	return
}

func scanTagGroup(rows *sql.Rows) (*TagGroup, error) {
	var id, createdAt, updatedAt string
	var protected int
	var tagsJSON []byte
	err := rows.Scan(&id, &tagsJSON, &protected, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return &TagGroup{
		ID:        id,
		Tags:      mustGetTags(tagsJSON),
		Protected: itob(protected),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
func itob(i int) bool {
	if i > 0 {
		return true
	}
	return false
}
func mustGetTags(data []byte) []string {
	var tags []string
	err := json.Unmarshal(data, &tags)
	util.Panic(err)
	return tags
}

// DB .
type DB struct {
	path string
	DB   *storm.DB
	Sess *session.Store

	// 只在 package database 外部使用锁，不在 package database 内部使用锁。
	sync.Mutex
}

// Open .
func (db *DB) Open(dbPath string) (err error) {
	if db.DB, err = storm.Open(dbPath); err != nil {
		return err
	}
	db.path = dbPath
	db.Sess = session.New(session.Config{
		Expiration: mustParseDuration(config.MaxAge),
		CookieName: cookieName,
	})
	err1 := db.createIndexes()
	err2 := db.initFirstID()
	err3 := db.initTotalSize()
	return util.WrapErrors(err1, err2, err3)
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	util.Panic(err)
	return d
}

// Close 只是 db.DB.Close(), 不清空 db 里的其它部分。
func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) mustBegin() storm.Node {
	tx, err := db.DB.Begin(true)
	util.Panic(err)
	return tx
}

// 创建 bucket 和索引
func (db *DB) createIndexes() error {
	err1 := db.DB.Init(&Note{})
	err2 := db.DB.Init(&Tag{})
	err3 := db.DB.Init(&TagGroup{})
	err4 := db.reIndex()
	return util.WrapErrors(err1, err2, err3, err4)
}

func (db *DB) reIndex() error {
	// 不知道为啥 TagGroup 的 index 经常出问题
	err1 := db.DB.ReIndex(&Note{})
	err2 := db.DB.ReIndex(&Tag{})
	err3 := db.DB.ReIndex(&TagGroup{})
	return util.WrapErrors(err1, err2, err3)
}

// Upgrade 将旧的历史版本系统（全文保存）升级至新的历史版本系统（只保存差异）。
func (db *DB) Upgrade() error {
	if db.noNeedToUpgrade() {
		return nil
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	var all []Note
	if err := tx.All(&all); err != nil {
		return err
	}
	for _, note := range all {
		histories, err := txNoteHistories(tx, note.ID)
		if err != nil {
			return err
		}

		// 加头加尾
		first := new(History)
		histories = append([]History{*first}, histories...)
		last := History{Contents: note.Contents}
		histories = append(histories, last)

		for i := 1; i < len(histories); i++ {
			a := histories[i-1].Contents
			b := histories[i].Contents
			patch, err := getUnifiedDiffString(a, b)
			if err != nil {
				return err
			}
			if patch != "" {
				note.Patches = append(note.Patches, patch)
			}
		}
		note.Contents = "" // 清空 Contents, 历史版本系统升级后废除 Contents
		err1 := tx.Save(&note)
		err2 := txIncreaseTotalSize(tx, note.Size) // 估算 size，不准确但问题不大
		if err := util.WrapErrors(err1, err2); err != nil {
			return err
		}
	}
	if err := tx.Drop("History"); err != nil {
		return err
	}
	return tx.Commit()
}

func txNoteHistories(tx storm.Node, noteID string) (histories []History, err error) {
	err = tx.Select(q.Eq("NoteID", noteID)).
		OrderBy("CreatedAt").Find(&histories)
	if err == storm.ErrNotFound {
		err = nil
	}
	return
}

func getUnifiedDiffString(a, b string) (string, error) {
	diff := difflib.LineDiffParams{
		A:        difflib.SplitLines(a),
		B:        difflib.SplitLines(b),
		FromFile: " ",
		ToFile:   " ",
	}
	return difflib.GetUnifiedDiffString(diff)
}

func (db *DB) noNeedToUpgrade() bool {
	var histories []History
	err := db.DB.All(&histories)
	if err == storm.ErrNotFound {
		return true
	}
	util.Panic(err)
	if len(histories) == 0 {
		return true
	}
	return false
}

// NewNote .
func (db *DB) NewNote(noteType model.NoteType) *Note {
	id := db.mustGetNextID()
	return model.NewNote(id.String(), noteType)
}

// Insert .
func (db *DB) Insert(note *Note) error {
	if err := db.checkTotalSize(note.Size); err != nil {
		return err
	}
	if err := db.checkExist(note.ID); err != nil {
		return err
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	err1 := tx.Save(note)
	err2 := saveTagGroup(tx, model.NewTagGroup(note.Tags))
	err3 := addTags(tx, note.Tags, note.ID)
	if err := util.WrapErrors(err1, err2, err3); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return db.increaseTotalSize(note.Size)
}

// SaveTagGroup .
func (db *DB) SaveTagGroup(tagGroup *TagGroup) error {
	return saveTagGroup(db.DB, tagGroup)
}

func saveTagGroup(tx storm.Node, tagGroup *TagGroup) (err error) {
	if len(tagGroup.Tags) < 2 {
		return
	}
	err = tx.Save(tagGroup)
	if err == storm.ErrAlreadyExists {
		if err = tx.One("Tags", tagGroup.Tags, tagGroup); err != nil {
			return
		}
		err = tx.UpdateField(tagGroup, "UpdatedAt", model.TimeNow())
	}
	return deleteOldTagGroup(tx)
}

func deleteOldTagGroup(tx storm.Node) (err error) {
	groups, err := notProtectedTagGroups(tx)
	if err != nil {
		return err
	}
	if len(groups) > settings.Config.TagGroupLimit {
		oldGroup := groups[0]
		err = tx.DeleteStruct(&oldGroup)
	}
	return
}

func notProtectedTagGroups(tx storm.Node) (groups []TagGroup, err error) {
	err = tx.Select(q.Eq("Protected", false)).OrderBy("UpdatedAt").Find(&groups)
	if err == storm.ErrNotFound {
		err = nil
	}
	return
}

// 检查 ID 冲突
func (db *DB) checkExist(id string) error {
	_, err := db.GetByID(id)
	if err == nil {
		return errors.New("id: " + id + " already exists")
	}
	return nil
}

func (db *DB2) GetByID(id string) (note Note, err error) {
	row := stmtGetNote.QueryRow(id)
	if note, err = scanNote(row); err != nil {
		return
	}

	tags, err := getTextArray(stmtGetTagNamesByNote, note.ID)
	if err != nil {
		return
	}
	note.Tags = tags

	patches, err := getTextArray(stmtGetPatchesByNote, note.ID)
	if err != nil {
		return
	}
	note.Patches = patches
	return
}

func refillPatches() {}

// GetByID .
func (db *DB) GetByID(id string) (note Note, err error) {
	err = db.DB.One("ID", id, &note)
	return note, err
}

func addTags(tx storm.Node, tags []string, noteID string) error {
	for _, tagName := range tags {
		tag := new(Tag)
		err := tx.One("Name", tagName, tag)
		if err != nil && err != storm.ErrNotFound {
			return err
		}

		// if not found, it's a new tag.
		if err == storm.ErrNotFound {
			aTag := model.NewTag(tagName, noteID)
			if err := tx.Save(aTag); err != nil {
				return err
			}
			continue
		}

		// if found (err == nil)
		tag.Add(noteID)
		if err := tx.Update(tag); err != nil {
			return err
		}
	}
	return nil
}

func deleteTags(tx storm.Node, tagsToDelete []string, noteID string) error {
	for _, tagName := range tagsToDelete {
		tag := new(Tag)
		if err := tx.One("Name", tagName, tag); err != nil {
			return fmt.Errorf("tag[%s] %w", tagName, err)
		}
		tag.Remove(noteID) // 每一个 tag 都与该 Note.ID 脱离关系
		if err := tx.UpdateField(tag, "NoteIDs", tag.NoteIDs); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB2) AllNotes() (notes []*Note, err error) {
	return db.getNotes(stmtGetNotes)
}

func (db *DB2) AllDeletedNotes() (notes []*Note, err error) {
	return db.getNotes(stmtGetDeletedNotes)
}

func (db *DB2) getNotes(stmtGet *Stmt) (notes []*Note, err error) {
	rows, err := stmtGet.Query()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, &note)
	}
	if err = rows.Err(); err != nil {
		return
	}
	for _, note := range notes {
		tags, err := getTextArray(stmtGetTagNamesByNote, note.ID)
		if err != nil {
			return nil, err
		}
		note.Tags = tags
	}
	return
}

func getTextArray(stmtGet *Stmt, arg string) (textArray []string, err error) {
	rows, err := stmtGet.Query(arg)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		text, err := scanText1(rows)
		if err != nil {
			return nil, err
		}
		textArray = append(textArray, text)
	}
	err = rows.Err()
	return
}

func scanText1(row Row) (text string, err error) {
	err = row.Scan(&text)
	return
}

func scanNote(row Row) (note Note, err error) {
	var deleted int
	err = row.Scan(
		&note.ID,
		&note.Type,
		&note.Title,
		&note.Size,
		&deleted,
		&note.RemindAt,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		return
	}
	note.Deleted = itob(deleted)
	return
}

// AllNotes .
func (db *DB) AllNotes() (notes []Note, err error) {
	err = db.DB.Select(q.Eq("Deleted", false)).
		OrderBy("UpdatedAt").Find(&notes)
	return
}

// AllDeletedNotes .
func (db *DB) AllDeletedNotes() (notes []Note, err error) {
	err = db.DB.Select(q.Eq("Deleted", true)).
		OrderBy("UpdatedAt").Find(&notes)
	return
}

// AllNotesWithDeleted .
func (db *DB) AllNotesWithDeleted() (notes []Note, err error) {
	err = db.DB.AllByIndex("UpdatedAt", &notes)
	return
}

// AllTags fetches all tags, sorted by "Name".
func (db *DB) AllTags() (tags []Tag, err error) {
	err = db.DB.AllByIndex("Name", &tags)
	return
}

// AllTagsByDate fetches all tags, sorted by "CreatedAt".
func (db *DB) AllTagsByDate() (tags []Tag, err error) {
	err = db.DB.AllByIndex("CreatedAt", &tags)
	return
}

// AllTagGroups fetches all tag-groups, sortd by "UpdatedAt".
func (db *DB) AllTagGroups() ([]TagGroup, error) {
	groups, err := txAllTagGroups(db.DB)
	if err == storm.ErrNotFound {
		db.reIndex()
		return txAllTagGroups(db.DB)
	}
	return groups, nil
}

func txAllTagGroups(tx storm.Node) (groups []TagGroup, err error) {
	err = tx.AllByIndex("UpdatedAt", &groups)
	return
}

// ChangeType 同时也可能需要修改标题。
func (db *DB) ChangeType(id string, noteType NoteType) error {
	note, err := db.GetByID(id)
	if err != nil {
		return err
	}
	note.Type = noteType
	if noteType == model.Markdown {
		note.SetTitle(note.Title)
	}
	return db.DB.Update(&note)
}

// UpdateTags .
func (db *DB) UpdateTags(id string, tags []string) error {
	note, err := db.GetByID(id)
	if err != nil {
		return err
	}
	tx := db.mustBegin()
	defer tx.Rollback()

	toAdd, toDelete := util.SliceDifference(tags, note.Tags)

	e1 := deleteTags(tx, toDelete, note.ID)
	e2 := addTags(tx, toAdd, note.ID)
	e3 := note.SetTags(tags)
	e4 := tx.UpdateField(&note, "Tags", note.Tags)
	e5 := saveTagGroup(tx, model.NewTagGroup(tags))

	if err := util.WrapErrors(e1, e2, e3, e4, e5); err != nil {
		return err
	}
	return tx.Commit()
}

// ResetAllTags .
func (db *DB) ResetAllTags() error {
	tx := db.mustBegin()
	defer tx.Rollback()

	var all []Note
	if err := tx.All(&all); err != nil {
		return err
	}
	for i := range all {
		if err := addTags(tx, all[i].Tags, all[i].ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetTag .
func (db *DB) GetTag(name string) (tag Tag, err error) {
	err = db.DB.One("Name", name, &tag)
	return
}

func (db *DB2) AddPatchSetTitle(id, patch, title string) (size int, err error) {
	var note Note
	size = len(patch)
	if note, err = db.GetByID(id); err != nil {
		return
	}
	if err = note.UpdateTitleSizeNow(title, size); err != nil {
		return
	}
	
	tx := db.mustBegin()
	defer tx.Rollback()

	stmt1 := mustPrepare(tx, stmt.InsertPatch)
	defer stmt1.Close()
	stmt2 := mustPrepare(tx, stmt.InsertNotePatch)
	defer stmt2.Close()

	if err = insertPatch(stmt1, stmt2, id, patch); err != nil {
		return
	}
	if _, err = tx.Exec(stmt.UpdateTitleSizeNow,
		note.Title, note.Size, note.UpdatedAt, note.ID); err != nil {
		return
	}
	err = tx.Commit()
	return
}

// AddPatchSetTitle .
func (db *DB) AddPatchSetTitle(id, patch, contents string) (int, error) {
	note, err := db.GetByID(id)
	if err != nil {
		return 0, err
	}
	if err := note.AddPatchNow(patch, contents); err != nil {
		return 0, err
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	err1 := tx.Update(&note)
	err2 := txCheckIncreaseTotalSize(tx, len(patch))

	if err := util.WrapErrors(err1, err2); err != nil {
		return 0, err
	}
	err = tx.Commit()
	return len(note.Patches), err
}

func txUnprotectedHistories(tx storm.Node, noteID string) (histories []History, err error) {
	err = tx.Select(q.Eq("NoteID", noteID), q.Eq("Protected", false)).
		OrderBy("CreatedAt").Find(&histories)
	if err == storm.ErrNotFound {
		err = nil
	}
	return
}

// SetTagGroupProtected .
func (db *DB) SetTagGroupProtected(groupID string, protected bool) error {
	return db.DB.UpdateField(
		&TagGroup{ID: groupID}, "Protected", protected)
}

// GetByTag returns notes without contents.
func (db *DB) GetByTag(name string) (notes []Note, err error) {
	tag, err := db.GetTag(name)
	if err != nil {
		return nil, fmt.Errorf("tag[%s] %w", name, err)
	}
	for i := range tag.NoteIDs {
		var note Note
		note, err = db.GetByID(tag.NoteIDs[i])
		if err != nil {
			return
		}
		note.Patches = nil
		notes = append(notes, note)
	}
	return
}

// RenameTag .
func (db *DB) RenameTag(oldName, newName string) error {
	_, err := db.GetTag(newName)
	if err != nil && err != storm.ErrNotFound {
		return fmt.Errorf("tag[%s] %w", newName, err)
	}
	if err == nil {
		return errors.New("标签名称 [" + newName + "] 已存在")
	}

	tag, err := db.GetTag(oldName)
	if err != nil {
		return fmt.Errorf("tag[%s] %w", oldName, err)
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	if err := renameTag(tx, tag, newName); err != nil {
		return err
	}
	return tx.Commit()
}

func renameTag(tx storm.Node, tag Tag, newName string) error {
	err1 := notesRenameTag(tx, tag, newName)
	err2 := tagGroupsRenameTag(tx, tag.Name, newName)
	err3 := tx.DeleteStruct(&tag)

	tag.Name = newName
	err4 := tx.Save(&tag)
	return util.WrapErrors(err1, err2, err3, err4)
}

func notesRenameTag(tx storm.Node, tag Tag, newName string) error {
	for _, noteID := range tag.NoteIDs {
		var note Note
		if err := tx.One("ID", noteID, &note); err != nil {
			return fmt.Errorf("id[%s] %w", noteID, err)
		}
		note.RenameTag(tag.Name, newName)
		if err := tx.UpdateField(&note, "Tags", note.Tags); err != nil {
			return err
		}
	}
	return nil
}

func tagGroupsRenameTag(tx storm.Node, oldName, newName string) error {
	groups, err := txAllTagGroups(tx)
	if err != nil {
		return err
	}
	for _, group := range groups {
		group.RenameTag(oldName, newName)
		if err := tx.UpdateField(&group, "Tags", group.Tags); err != nil {
			return err
		}
	}
	return nil
}

// SearchTagGroup 通过标签组搜索笔记。
// 如果其中一个标签不存在，会返回错误，另外一种处理方式是忽略找不到的标签。
// 但我选择了返回错误，因为本项目的设计思想之一是 informational(更多信息)。
func (db *DB) SearchTagGroup(tags []string) ([]Note, error) {
	var idGroups []*Set
	for i := range tags {
		var tag Tag
		if err := db.DB.One("Name", tags[i], &tag); err != nil {
			return nil, fmt.Errorf("Tag[%s] %w", tags[i], err)
		}
		idGroups = append(idGroups, stringset.NewSet(tag.NoteIDs))
	}
	noteIDs := stringset.Intersect(idGroups).Slice()
	return db.getByIDs(noteIDs)
}

func (db *DB) getByIDs(noteIDs []string) ([]Note, error) {
	var notes []Note
	err := db.DB.Select(q.In("ID", noteIDs)).
		OrderBy("UpdatedAt").Find(&notes)
	if err == storm.ErrNotFound {
		err = nil
	}
	return notes, err
}

// SearchTitle by regular expression.
func (db *DB) SearchTitle(pattern string) ([]Note, error) {
	var notes []Note
	err := db.DB.Select(q.Re("Title", pattern)).
		OrderBy("UpdatedAt").Find(&notes)
	if err == storm.ErrNotFound {
		err = nil
	}
	return notes, err
}

// SetNoteDeleted .
func (db *DB) SetNoteDeleted(id string, deleted bool) error {
	note, err := db.GetByID(id)
	if err != nil {
		return err
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	var err1 error
	if deleted {
		err1 = deleteTags(tx, note.Tags, note.ID)
	} else {
		err1 = addTags(tx, note.Tags, note.ID)
	}
	err2 := tx.UpdateField(&note, "Deleted", deleted)
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteNoteForever .
func (db *DB) DeleteNoteForever(id string) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	if err := txDeleteOneNote(tx, id); err != nil {
		return err
	}
	return tx.Commit()
}

func txDeleteOneNote(tx storm.Node, id string) error {
	var note Note
	err1 := tx.One("ID", id, &note)
	err2 := tx.DeleteStruct(&note)
	err3 := txIncreaseTotalSize(tx, -note.Size)
	return util.WrapErrors(err1, err2, err3)
}

// DeleteTag .
func (db *DB) DeleteTag(name string) error {
	tag, err := db.GetTag(name)
	if err != nil {
		return fmt.Errorf("tag[%s] %w", name, err)
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	err1 := notesDeleteTag(tx, tag)
	err2 := tx.DeleteStruct(&tag)
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return tx.Commit()
}

func notesDeleteTag(tx storm.Node, tag Tag) error {
	for _, noteID := range tag.NoteIDs {
		var note Note
		if err := tx.One("ID", noteID, &note); err != nil {
			return fmt.Errorf("id[%s] %w", noteID, err)
		}
		note.DeleteTag(tag.Name)
		if err := tx.UpdateField(&note, "Tags", note.Tags); err != nil {
			return err
		}
	}
	return nil
}

// DeleteNoteHistory .
func (db *DB) DeleteNoteHistory(noteID string) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	query := tx.Select(q.Eq("NoteID", noteID), q.Eq("Protected", false))
	if err := txDeleteHistories(tx, query); err != nil {
		return err
	}
	return tx.Commit()
}

func txDeleteHistories(tx storm.Node, query storm.Query) error {
	var (
		size      int
		histories []History
	)
	err1 := query.Find(&histories)
	if err1 == storm.ErrNotFound {
		return nil
	}
	for i := range histories {
		size += histories[i].Size
	}
	err2 := txIncreaseTotalSize(tx, -size)
	err3 := query.Delete(&History{})
	return util.WrapErrors(err1, err2, err3)
}
