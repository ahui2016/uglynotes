package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/session"
	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/stmt"
	"github.com/ahui2016/uglynotes/stringset"
	"github.com/ahui2016/uglynotes/tagset"
	"github.com/ahui2016/uglynotes/util"
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

type TX interface {
	Exec(string, ...interface{}) (sql.Result, error)
	QueryRow(string, ...interface{}) *sql.Row
	Prepare(string) (*Stmt, error)
}

type Row interface {
	Scan(...interface{}) error
}

type DB struct {
	path string
	DB   *sql.DB
	Sess *session.Manager
	sync.Mutex
}

func (db *DB) Open(dbPath string) (err error) {
	if db.DB, err = sql.Open("sqlite3", dbPath+"?_fk=1"); err != nil {
		return
	}
	if err = db.Exec(stmt.CreateTables); err != nil {
		return
	}
	db.path = dbPath
	maxAge := mustParseDuration(config.MaxAge)
	db.Sess = session.NewManager(cookieName, int(maxAge))
	err1 := initFirstID(db.DB)
	err2 := initTotalSize(db.DB)
	return util.WrapErrors(err1, err2)
}
func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) mustBegin() *sql.Tx {
	tx, err := db.DB.Begin()
	util.Panic(err)
	return tx
}

func mustPrepare(tx TX, query string) *Stmt {
	stmt, err := tx.Prepare(query)
	util.Panic(err)
	return stmt
}

func (db *DB) Exec(query string, args ...interface{}) (err error) {
	_, err = db.DB.Exec(query, args...)
	return
}

func exec(aStmt *Stmt, args ...interface{}) (err error) {
	_, err = aStmt.Exec(args...)
	return
}

func (db *DB) ImportNotes(notes []Note) (err error) {
	tx := db.mustBegin()
	defer tx.Rollback()

	for _, note := range notes {
		if err = addNoteTagPatch(tx, &note); err != nil {
			return
		}
	}
	if _, err = resetCurrentID(tx); err != nil {
		return
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

func (db *DB) ResetCurrentID() (newid string, err error) {
	return resetCurrentID(db.DB)
}

func resetCurrentID(tx TX) (newid string, err error) {
	strID, err := getLastNoteID(tx)
	if err != nil {
		return
	}
	id, err := model.ParseID(strID)
	if err != nil {
		return
	}
	newid = id.Increase().String()
	return newid, setCurrentID(tx, newid)
}

func getLastNoteID(tx TX) (id string, err error) {
	row := tx.QueryRow(stmt.GetLastNoteID)
	err = row.Scan(&id)
	return
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

func (db *DB) AddTagGroup(group *TagGroup) error {
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

func (db *DB) GetTagByID(id string) (tag Tag, err error) {
	row := db.DB.QueryRow(stmt.GetTag, id)
	err = row.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	return
}

func (db *DB) GetTagByName(name string) (tag Tag, err error) {
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

func (db *DB) AllTagGroups() (groups []TagGroup, err error) {
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

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	util.Panic(err)
	return d
}

// NewNote .
func (db *DB) NewNote(
	title, patch string, noteType NoteType, tagNames []string) (*Note, error) {
	id := db.mustGetNextID()
	return model.NewNote(id, title, patch, noteType, tagNames)
}

// Insert .
func (db *DB) Insert(note *Note) error {
	tx := db.mustBegin()
	defer tx.Rollback()
	if err := addNoteTagPatch(tx, note); err != nil {
		return err
	}
	return tx.Commit()
}

// getNote gets a note without tags and patches.
func (db *DB) getNote(id string) (note Note, err error) {
	row := db.DB.QueryRow(stmt.GetNote, id)
	return scanNote(row)
}

func (db *DB) GetByID(id string) (note Note, err error) {
	if note, err = db.getNote(id); err != nil {
		return
	}

	tags, err := db.getSimpleTagsByNote(id)
	if err != nil {
		return
	}
	note.Tags = tags

	stmtGetPatchesByNote := mustPrepare(db.DB, stmt.GetPatchesByNote)
	defer stmtGetPatchesByNote.Close()
	patches, err := getTextArray(stmtGetPatchesByNote, id)
	if err != nil {
		return
	}
	note.Patches = patches
	return
}

func (db *DB) getSimpleTagsByNote(id string) ([]tagset.Tag, error) {
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

func (db *DB) AllNotes() (notes []*Note, err error) {
	stmtGetNotes := mustPrepare(db.DB, stmt.GetNotes)
	defer stmtGetNotes.Close()
	return db.getNotes(stmtGetNotes)
}

func (db *DB) ExportAllNotes() (notes []Note, err error) {
	noteIDs, err := db.getAllNoteIDs()
	if err != nil {
		return nil, err
	}
	for i := range noteIDs {
		note, err := db.GetByID(noteIDs[i])
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return
}

func (db *DB) getAllNoteIDs() (noteIDs []string, err error) {
	rows, err := db.DB.Query(stmt.GetAllNoteIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		noteIDs = append(noteIDs, id)
	}
	err = rows.Err()
	return
}

func (db *DB) AllDeletedNotes() (notes []*Note, err error) {
	stmtGetDeletedNotes := mustPrepare(db.DB, stmt.GetDeletedNotes)
	defer stmtGetDeletedNotes.Close()
	return db.getNotes(stmtGetDeletedNotes)
}

func (db *DB) getNotes(stmtGet *Stmt) (notes []*Note, err error) {
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

func (db *DB) fillSimpleTags(notes []*Note) error {
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

// AllTags fetches all tags, sorted by "Name".
func (db *DB) AllTagsByName() (tags []Tag, err error) {
	return db.getAllTags(stmt.AllTagsByName)
}

// AllTagsByDate fetches all tags, sorted by "CreatedAt".
func (db *DB) AllTagsByDate() (tags []Tag, err error) {
	return db.getAllTags(stmt.AllTagsByDate)
}

func (db *DB) getAllTags(stmtGet string) (tags []Tag, err error) {
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

// ChangeType 同时也可能需要修改标题。
func (db *DB) ChangeType(id string, noteType NoteType) error {
	note, err := db.getNote(id)
	if err != nil {
		return err
	}
	note.Type = noteType
	if noteType == model.Markdown {
		note.SetTitle(note.Title)
		return db.Exec(stmt.SetTypeTitle, noteType, note.Title, id)
	}
	return db.Exec(stmt.ChangeNoteType, noteType, id)
}

// UpdateTags .
func (db *DB) UpdateTags(id string, tagNames []string) error {
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

func (db *DB) AddPatchSetTitle(id, patch, title string) (
	count int, err error) {
	var note Note
	size := len(patch)
	if note, err = db.getNote(id); err != nil {
		return
	}
	if err = note.UpdateTitleSizeNow(title, size); err != nil {
		return
	}

	tx := db.mustBegin()
	defer tx.Rollback()

	err1 := addPatches(tx, id, []string{patch})
	_, err2 := tx.Exec(stmt.UpdateTitleSizeNow,
		note.Title, note.Size, note.UpdatedAt, note.ID)
	err3 := increaseTotalSize(tx, size)
	row := tx.QueryRow(stmt.CountPatches, note.ID)
	err4 := row.Scan(&count)
	if err = util.WrapErrors(err1, err2, err3, err4); err != nil {
		return
	}
	err = tx.Commit()
	return
}

// SetTagGroupProtected .
func (db *DB) SetTagGroupProtected(groupID string, protected bool) error {
	return db.Exec(stmt.SetTagGroupProtected, protected, groupID)
}

func (db *DB) GetNotesByTagID(tagID string) ([]*Note, error) {
	noteIDs, err := db.getNoteIDs(stmt.GetNotesByTagID, tagID)
	if err != nil {
		return nil, err
	}
	return db.getNotesByIDs(noteIDs)
}

func (db *DB) getNotesByIDs(noteIDs []string) (notes []*Note, err error) {
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

func (db *DB) getNoteIDs(stmtGet, arg string) (noteIDs []string, err error) {
	rows, err := db.DB.Query(stmtGet, arg)
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
	if err = rows.Err(); err != nil {
		return
	}
	if len(noteIDs) == 0 {
		err = errors.New("no notes related to " + arg)
	}
	return
}

// RenameTag .
func (db *DB) RenameTag(id, newName string) error {
	return db.Exec(stmt.RenameTag, newName, id)
}

// SearchTagGroup 通过标签组搜索笔记。
func (db *DB) SearchTagGroup(tagNames []string) ([]*Note, error) {
	noteIDs, err := db.getNoteIDsByTagNames(tagNames)
	if err != nil {
		return nil, err
	}
	return db.getNotesByIDs(noteIDs)
}

func (db *DB) getNoteIDsByTagNames(tagNames []string) ([]string, error) {
	var idSets []*Set
	for _, tagName := range tagNames {
		ids, err := db.getNoteIDs(stmt.GetNotesByTagName, tagName)
		if err != nil {
			return nil, err
		}
		idSets = append(idSets, stringset.From(ids))
	}
	return stringset.Intersect(idSets).Slice(), nil
}

func (db *DB) SearchTitle(pattern string) (notes []*Note, err error) {
	rows, err := db.DB.Query(stmt.SearchNoteTitle, "%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, &note)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	db.fillSimpleTags(notes)
	return
}

func (db *DB) SetNoteDeleted(id string, deleted bool) error {
	return db.Exec(stmt.SetNoteDeleted, deleted, id)
}

func (db *DB) DeleteNoteForever(id string) error {
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

func (db *DB) getNoteSize(id string) (size int, err error) {
	row := db.DB.QueryRow(stmt.GetNoteSize, id)
	err = row.Scan(&size)
	return
}

// DeleteTag .
func (db *DB) DeleteTag(id string) error {
	return db.Exec(stmt.DeleteTag, id)
}
