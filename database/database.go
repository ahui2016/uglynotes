package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/stringset"
	"github.com/ahui2016/uglynotes/util"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/gofiber/fiber/v2/middleware/session"
)

const cookieName = "uglynotesCookie"

type (
	Note       = model.Note
	NoteType   = model.NoteType
	History    = model.History
	Tag        = model.Tag
	TagGroup   = model.TagGroup
	IncreaseID = model.IncreaseID
	Set        = stringset.Set
)

// DB .
type DB struct {
	path     string
	capacity int
	DB       *storm.DB
	Sess     *session.Store

	// 只在 package database 外部使用锁，不在 package database 内部使用锁。
	sync.Mutex
}

// Open .
func (db *DB) Open(dbPath string) (err error) {
	if db.DB, err = storm.Open(dbPath); err != nil {
		return err
	}
	db.path = dbPath
	db.capacity = settings.DatabaseCapacity
	db.Sess = session.New(session.Config{
		Expiration: settings.MaxAge,
		CookieName: cookieName,
	})
	err1 := db.createIndexes()
	err2 := db.initFirstID()
	err3 := db.initCapacity()
	err4 := db.initTotalSize()
	return util.WrapErrors(err1, err2, err3, err4)
}

// Close 只是 db.DB.Close(), 不清空 db 里的其它部分。
func (db *DB) Close() error {
	return db.DB.Close()
}

// 创建 bucket 和索引
func (db *DB) createIndexes() error {
	err1 := db.DB.Init(&Note{})
	err2 := db.DB.Init(&History{})
	err3 := db.DB.Init(&Tag{})
	err4 := db.DB.Init(&TagGroup{})
	return util.WrapErrors(err1, err2, err3, err4)
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

	tx, err := db.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := tx.Save(note); err != nil {
		return err
	}
	if err := saveTagGroup(tx, model.NewTagGroup(note.Tags)); err != nil {
		return err
	}
	if err := addTags(tx, note.Tags, note.ID); err != nil {
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
	if len(groups) > settings.TagGroupLimit {
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

// AllNotes fetches all notes, sorted by "UpdatedAt".
func (db *DB) AllNotes() (notes []Note, err error) {
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
func (db *DB) AllTagGroups() (groups []TagGroup, err error) {
	return txAllTagGroups(db.DB)
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
	// 这里可以优化性能，暂时先不优化。
	note.SetContents(note.Contents)
	return db.DB.Update(&note)
}

// UpdateTags .
func (db *DB) UpdateTags(id string, tags []string) error {
	note, err := db.GetByID(id)
	if err != nil {
		return err
	}
	tx, err := db.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	toAdd, toDelete := util.SliceDifference(tags, note.Tags)

	// 删除标签（从 tag.NoteIDs 里删除 id）
	if err := deleteTags(tx, toDelete, note.ID); err != nil {
		return err
	}

	// 添加标签（将 id 添加到 tag.NoteIDs 里）
	if err := addTags(tx, toAdd, note.ID); err != nil {
		return err
	}

	// 更新标签组
	if err := saveTagGroup(tx, model.NewTagGroup(tags)); err != nil {
		return err
	}

	// 最后更新 note.Tags
	note.SetTags(tags)
	if err := tx.UpdateField(&note, "Tags", note.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

// GetTag .
func (db *DB) GetTag(name string) (tag Tag, err error) {
	err = db.DB.One("Name", name, &tag)
	return
}

// GetHistory .
func (db *DB) GetHistory(id string) (history History, err error) {
	err = db.DB.One("ID", id, &history)
	return
}

// UpdateNoteContents .
func (db *DB) UpdateNoteContents(id, contents string) (historyID string, err error) {
	note, err := db.GetByID(id)
	if err != nil {
		return "", err
	}
	history := model.NewHistory(note.Contents, id)
	note.SetContentsNow(contents)

	tx, err := db.DB.Begin(true)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if err := addHistory(tx, note, history); err != nil {
		return "", err
	}
	if err := txIncreaseTotalSize(tx, note.Size); err != nil {
		return "", err
	}
	err = tx.Commit()
	return history.ID, err
}

// addHistory 添加新 history, 同时可能需要删除旧的 history.
func addHistory(tx storm.Node, note Note, history *History) error {
	histories, err := txUnprotectedHistories(tx, note.ID)
	if err != nil {
		return err
	}
	if len(histories) > settings.HistoryLimit {
		oldHistory := histories[0]
		if err := deleteHistory(tx, oldHistory); err != nil {
			return err
		}
	}

	// 原笔记的体积转移至历史，因此只新增新笔记的体积。
	if err := txCheckTotalSize(tx, note.Size); err != nil {
		return err
	}
	if err := tx.Save(history); err != nil {
		return err
	}
	return tx.Update(&note)
}

func deleteHistory(tx storm.Node, oldHistory History) error {
	if err := tx.DeleteStruct(&oldHistory); err != nil {
		return err
	}
	return txIncreaseTotalSize(tx, -oldHistory.Size)
}

// NoteHistories .
func (db *DB) NoteHistories(noteID string) (histories []History, err error) {
	err = db.DB.Select(q.Eq("NoteID", noteID)).
		OrderBy("CreatedAt").Find(&histories)
	if err == storm.ErrNotFound {
		err = nil
	}
	return
}

func txUnprotectedHistories(tx storm.Node, noteID string) (histories []History, err error) {
	err = tx.Select(q.Eq("NoteID", noteID), q.Eq("Protected", false)).
		OrderBy("CreatedAt").Find(&histories)
	if err == storm.ErrNotFound {
		err = nil
	}
	return
}

// SetProtected .
func (db *DB) SetProtected(historyID string, protected bool) error {
	return db.DB.UpdateField(
		&History{ID: historyID}, "Protected", protected)
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
		note.Contents = ""
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

	tx, err := db.DB.Begin(true)
	if err != nil {
		return err
	}
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

// DeleteNote .
func (db *DB) DeleteNote(id string) error {
	note, err := db.GetByID(id)
	if err != nil {
		return err
	}

	tx, err := db.DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err1 := deleteTags(tx, note.Tags, note.ID)
	err2 := tx.DeleteStruct(&note)
	if err := util.WrapErrors(err1, err2); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteTag .
func (db *DB) DeleteTag(name string) error {
	tag, err := db.GetTag(name)
	if err != nil {
		return fmt.Errorf("tag[%s] %w", name, err)
	}

	tx, err := db.DB.Begin(true)
	if err != nil {
		return err
	}
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
	return db.DB.Select(q.Eq("NoteID", noteID), q.Eq("Protected", false)).
		Delete(&History{})
}
