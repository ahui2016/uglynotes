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
	"github.com/ahui2016/uglynotes/tagset"
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
	Tag        = model.Tag
	TagGroup   = model.TagGroup
	IncreaseID = model.IncreaseID
	Set        = stringset.Set
	Stmt       = sql.Stmt
)

var stmtGetTagByNote, stmtGetNotes, stmtGetDeletedNotes,
	stmtGetPatchesByNote, stmtSetNoteDeleted, stmtGetNoteSize,
	stmtChangeNoteType, stmtSetTypeTitle *Stmt

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
		return
	}
	if err = db.Exec(stmt.CreateTables); err != nil {
		return
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
	stmtGetNotes = mustPrepare(db.DB, stmt.GetNotes)
	stmtGetDeletedNotes = mustPrepare(db.DB, stmt.GetDeletedNotes)
	stmtGetPatchesByNote = mustPrepare(db.DB, stmt.GetPatchesByNote)
	stmtSetNoteDeleted = mustPrepare(db.DB, stmt.SetNoteDeleted)
	stmtGetNoteSize = mustPrepare(db.DB, stmt.GetNoteSize)
	stmtChangeNoteType = mustPrepare(db.DB, stmt.ChangeNoteType)
	stmtSetTypeTitle = mustPrepare(db.DB, stmt.SetTypeTitle)
}

func closeStatements() {
	stmtGetNotes.Close()
	stmtGetDeletedNotes.Close()
	stmtGetPatchesByNote.Close()
	stmtSetNoteDeleted.Close()
	stmtGetNoteSize.Close()
	stmtChangeNoteType.Close()
	stmtSetTypeTitle.Close()
}

func (db *DB2) Exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
}

func exec(aStmt *Stmt, args ...interface{}) (err error) {
	_, err = aStmt.Exec(args...)
	return
}

func (db *DB2) ImportNotes(notes []Note) (err error) {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, note := range notes {
		if err = addNoteTagPatch(tx, &note); err != nil {
			return
		}
	}
	return tx.Commit()
}

func addNoteTagPatch(tx TX, note *Note) (err error) {
	if err = insertNote(tx, note); err != nil {
		return fmt.Errorf("insertNote: %v", err)
	}
	group := model.NewTagGroup(tagset.ToNames(note.Tags))
	if err = addTagGroup(tx, group); err != nil {
		return fmt.Errorf("addTagGroup: %v", err)
	}
	if err = addTagNames(tx, tagset.ToNames(note.Tags), note.ID); err != nil {
		return fmt.Errorf("addTagNames: %v", err)
	}
	if err = addPatches(tx, note.ID, note.Patches); err != nil {
		return fmt.Errorf("addPatches: %v", err)
	}
	return increaseTotalSize(tx, note.Size)
}

func insertNote(tx TX, note *Note) (err error) {
	_, err = tx.Exec(stmt.InsertNote,
		note.ID,
		note.Type,
		note.Title,
		note.Size,
		note.Deleted,
		"",
		note.CreatedAt,
		note.UpdatedAt,
	)
	return
}

func (db *DB2) AddTagGroup(group *TagGroup) error {
	return addTagGroup(db.DB, group)
}

func addTagGroup(tx TX, group *TagGroup) error {
	stmtGetTagGroupID := mustPrepare(tx, stmt.GetTagGroupID)
	defer stmtGetTagGroupID.Close()
	stmtInsertTagGroup := mustPrepare(tx, stmt.InsertTagGroup)
	defer stmtInsertTagGroup.Close()
	stmtUpdateTagGroupNow := mustPrepare(tx, stmt.UpdateTagGroupNow)
	defer stmtUpdateTagGroupNow.Close()

	groupID, err := getTagGroupID(stmtGetTagGroupID, group.Tags)

	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		err = exec(stmtInsertTagGroup,
			group.ID,
			util.MustMarshal(group.Tags),
			group.Protected,
			group.CreatedAt,
			group.UpdatedAt)
	} else {
		// err == nil
		err = updateNow(stmtUpdateTagGroupNow, groupID)
	}
	if err != nil {
		return err
	}
	return deleteOldTagGroup(tx)
}

func deleteOldTagGroup(tx TX) (err error) {
	var count int
	row := tx.QueryRow(stmt.TagGroupCount)
	if err = row.Scan(&count); err != nil {
		return
	}
	if count < settings.Config.TagGroupLimit {
		return
	}
	var groupID string
	row = tx.QueryRow(stmt.LastTagGroup)
	if err = row.Scan(&groupID); err != nil {
		return
	}
	_, err = tx.Exec(stmt.DeleteTagGroup, groupID)
	return
}

func updateNow(stmtUpdate *Stmt, arg string) error {
	return exec(stmtUpdate, model.TimeNow(), arg)
}

func getTagGroupID(stmtGet *Stmt, tags []string) (string, error) {
	return getText1(stmtGet, util.MustMarshal(tags))
}

func (db *DB2) GetTagByID(id string) (tag Tag, err error) {
	row := db.DB.QueryRow(stmt.GetTag, id)
	err = row.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	return
}

func (db *DB2) GetTagByName(name string) (tag Tag, err error) {
	row := db.DB.QueryRow(stmt.GetTagByName, name)
	err = row.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	return
}

// getText1 gets a text value by one argument.
func getText1(stmtGet *Stmt, arg interface{}) (text string, err error) {
	row := stmtGet.QueryRow(arg)
	err = row.Scan(&text)
	return
}

func addTagNames(tx TX, tagNames []string, noteID string) (err error) {
	stmtGetTagID := mustPrepare(tx, stmt.GetTagID)
	defer stmtGetTagID.Close()
	stmtInsertTag := mustPrepare(tx, stmt.InsertTag)
	defer stmtInsertTag.Close()
	stmtInsertNoteTag := mustPrepare(tx, stmt.InsertNoteTag)
	defer stmtInsertNoteTag.Close()

	for _, tagName := range tagNames {
		err = addTag(
			stmtGetTagID, stmtInsertTag, stmtInsertNoteTag, noteID, tagName)
		if err != nil {
			return
		}
	}
	return
}

func addTag(stmtGet, stmtAdd, stmt3 *Stmt, noteID string, name string) error {
	tagID, err := getText1(stmtGet, name)
	if err == sql.ErrNoRows {
		tagID = model.RandomID()
		err = exec(stmtAdd, tagID, name, model.TimeNow())
	}
	if err != nil {
		return err
	}
	return exec(stmt3, noteID, tagID)
}

func addPatches(tx TX, noteID string, patches []string) (
	err error) {
	stmtInsertPatch := mustPrepare(tx, stmt.InsertPatch)
	defer stmtInsertPatch.Close()
	stmtInsertNotePatch := mustPrepare(tx, stmt.InsertNotePatch)
	defer stmtInsertNotePatch.Close()

	for _, diff := range patches {
		if err = addPatch(
			stmtInsertPatch, stmtInsertNotePatch, noteID, diff); err != nil {
			return
		}
	}
	return
}

func addPatch(stmt1, stmt2 *Stmt, noteID, diff string) error {
	patchID := model.NextTimeID()
	err1 := exec(stmt1, patchID, diff)
	err2 := exec(stmt2, noteID, patchID)
	return util.WrapErrors(err1, err2)
}

func (db *DB2) FillGroups(groups []TagGroup) error {
	questions := make([]string, 0, len(groups))
	values := make([]interface{}, 0, len(groups)*5)
	for _, group := range groups {
		questions = append(questions, "(?,?,?,?,?)")
		values = append(values, group.ID)
		values = append(values, util.MustMarshal(group.Tags))
		values = append(values, group.Protected)
		values = append(values, group.CreatedAt)
		values = append(values, group.UpdatedAt)
	}
	stmt := fmt.Sprintf(
		"INSERT INTO taggroup (id, tags, protected, created_at, updated_at) VALUES %s",
		strings.Join(questions, ","))
	return db.Exec(stmt, values...)
}
func (db *DB2) DropTagGroup() error {
	return db.Exec("DROP TABLE IF EXISTS taggroup")
}

func (db *DB2) AllTagGroups() (groups []TagGroup, err error) {
	rows, err := db.DB.Query(stmt.AllTagGroups)
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
func mustGetTags(data []byte) []string {
	var tags []string
	err := json.Unmarshal(data, &tags)
	util.Panic(err)
	return tags
}

func itob(i int) (b bool) {
	if i > 0 {
		b = true
	}
	return
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

func getUnifiedDiffString(a, b string) (string, error) {
	diff := difflib.LineDiffParams{
		A:        difflib.SplitLines(a),
		B:        difflib.SplitLines(b),
		FromFile: " ",
		ToFile:   " ",
	}
	return difflib.GetUnifiedDiffString(diff)
}

// NewNote .
func (db *DB2) NewNote(
	title, patch string, noteType NoteType, tagNames []string) (*Note, error) {
	id := db.mustGetNextID()
	return model.NewNote(id, title, patch, noteType, tagNames)
}

// Insert .
func (db *DB2) Insert(note *Note) error {
	tx := db.mustBegin()
	defer tx.Rollback()
	if err := addNoteTagPatch(tx, note); err != nil {
		return err
	}
	return tx.Commit()
}

// 检查 ID 冲突
func (db *DB) checkExist(id string) error {
	_, err := db.GetByID(id)
	if err == nil {
		return errors.New("id: " + id + " already exists")
	}
	return nil
}

// getNote gets a note without tags and patches.
func (db *DB2) getNote(id string) (note Note, err error) {
	row := db.DB.QueryRow(stmt.GetNote, id)
	return scanNote(row)
}

func (db *DB2) GetByID(id string) (note Note, err error) {
	if note, err = db.getNote(id); err != nil {
		return
	}

	tags, err := db.getSimpleTagsByNote(id)
	if err != nil {
		return
	}
	note.Tags = tags

	patches, err := getTextArray(stmtGetPatchesByNote, id)
	if err != nil {
		return
	}
	note.Patches = patches
	return
}

func (db *DB2) getSimpleTagsByNote(id string) ([]tagset.Tag, error) {
	rows, err := db.DB.Query(stmt.GetTagsByNote, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSimpleTags(rows)
}

func scanSimpleTags(rows *sql.Rows) (tags []tagset.Tag, err error) {
	for rows.Next() {
		var tag tagset.Tag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	return
}

// GetByID .
func (db *DB) GetByID(id string) (note Note, err error) {
	err = db.DB.One("ID", id, &note)
	return note, err
}

func deleteTags(tx TX, tagsToDelete []string, noteID string) error {
	stmtGetTagID := mustPrepare(tx, stmt.GetTagID)
	defer stmtGetTagID.Close()

	for _, tagName := range tagsToDelete {
		tagID, err := getText1(stmtGetTagID, tagName)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(stmt.DeleteTags, noteID, tagID); err != nil {
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
	db.fillSimpleTags(notes)
	return
}

func (db *DB2) fillSimpleTags(notes []*Note) error {
	for _, note := range notes {
		tags, err := db.getSimpleTagsByNote(note.ID)
		if err != nil {
			return err
		}
		note.Tags = tags
	}
	return nil
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
	err = row.Scan(
		&note.ID,
		&note.Type,
		&note.Title,
		&note.Size,
		&note.Deleted,
		&note.RemindAt,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
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
func (db *DB2) AllTagsByName() (tags []Tag, err error) {
	return db.getAllTags(stmt.AllTagsByName)
}

// AllTagsByDate fetches all tags, sorted by "CreatedAt".
func (db *DB2) AllTagsByDate() (tags []Tag, err error) {
	return db.getAllTags(stmt.AllTagsByDate)
}

func (db *DB2) getAllTags(stmtGet string) (tags []Tag, err error) {
	rows, err := db.DB.Query(stmtGet)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		err = rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.Count)
		if err != nil {
			return
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
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
func (db *DB2) ChangeType(id string, noteType NoteType) error {
	note, err := db.getNote(id)
	if err != nil {
		return err
	}
	note.Type = noteType
	if noteType == model.Markdown {
		note.SetTitle(note.Title)
		return exec(stmtSetTypeTitle, noteType, note.Title, id)
	}
	return exec(stmtChangeNoteType, noteType, id)
}

// UpdateTags .
func (db *DB2) UpdateTags(id string, tagNames []string) error {
	oldTags, err := db.getSimpleTagsByNote(id)
	if err != nil {
		return err
	}
	oldTagNames := tagset.ToNames(oldTags)
	newTagNames := stringset.UniqueSort(tagNames)
	toAdd, toDelete := util.SliceDifference(newTagNames, oldTagNames)

	tx := db.mustBegin()
	defer tx.Rollback()

	e1 := deleteTags(tx, toDelete, id)
	e2 := addTagNames(tx, toAdd, id)
	e3 := addTagGroup(tx, model.NewTagGroup(newTagNames))

	if err := util.WrapErrors(e1, e2, e3); err != nil {
		return err
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
	if note, err = db.getNote(id); err != nil {
		return
	}
	if err = note.UpdateTitleSizeNow(title, size); err != nil {
		return
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	if err = addPatches(tx, id, []string{patch}); err != nil {
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

// SetTagGroupProtected .
func (db *DB2) SetTagGroupProtected(groupID string, protected bool) error {
	return db.Exec(stmt.SetTagGroupProtected, protected, groupID)
}

func (db *DB2) GetNotesByTag(tagID string) (notes []*Note, err error) {
	noteIDs, err := db.getNoteIDs(tagID)
	if err != nil {
		return nil, fmt.Errorf("tag id[%s] %w", tagID, err)
	}

	// 这里改成批量查询或改用复杂的 sql 可优化性能。
	for _, id := range noteIDs {
		note, err := db.getNote(id)
		if err != nil {
			return nil, err
		}
		notes = append(notes, &note)
	}
	db.fillSimpleTags(notes)
	return
}

func (db *DB2) getNoteIDs(tagID string) (noteIDs []string, err error) {
	rows, err := db.DB.Query(stmt.GetNotesByTag, tagID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return
		}
		noteIDs = append(noteIDs, id)
	}
	err = rows.Err()
	return
}

// RenameTag .
func (db *DB2) RenameTag(id, newName string) error {
	return db.Exec(stmt.RenameTag, newName, id)
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
		// idGroups = append(idGroups, stringset.NewSet(tag.NoteIDs))
		idGroups = append(idGroups, stringset.NewSet([]string{}))
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

func (db *DB2) SetNoteDeleted(id string, deleted bool) error {
	return exec(stmtSetNoteDeleted, deleted, id)
}

func (db *DB2) DeleteNoteForever(id string) error {
	tx := db.mustBegin()
	defer tx.Rollback()

	_, err1 := tx.Exec(stmt.DeleteNote, id)
	_, err2 := tx.Exec(stmt.DeleteTagsByNote, id)
	size, err3 := db.getNoteSize(id)
	err4 := increaseTotalSize(tx, -size)

	if err := util.WrapErrors(err1, err2, err3, err4); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB2) getNoteSize(id string) (size int, err error) {
	row := stmtGetNoteSize.QueryRow(id)
	err = row.Scan(&size)
	return
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
func (db *DB2) DeleteTag(id string) error {
	return db.Exec(stmt.DeleteTag, id)
}
