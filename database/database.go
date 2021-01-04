package database

import (
	"errors"
	"sync"
	"time"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/util"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/gofiber/fiber/v2/middleware/session"
)

const cookieName = "uglynotesCookie"

// historyLimit 限制每篇笔记可保留的历史上限。
const historyLimit = 3

type (
	Note       = model.Note
	NoteType   = model.NoteType
	History    = model.History
	Tag        = model.Tag
	IncreaseID = model.IncreaseID
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
func (db *DB) Open(maxAge time.Duration, cap int, dbPath string) (err error) {
	if db.DB, err = storm.Open(dbPath); err != nil {
		return err
	}
	db.path = dbPath
	db.capacity = cap
	db.Sess = session.New(session.Config{
		Expiration: maxAge,
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
	return util.WrapErrors(err1, err2, err3)
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
	if err := addTags(tx, note.Tags, note.ID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return db.increaseTotalSize(note.Size)
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
			return err
		}
		tag.Remove(noteID) // 每一个 tag 都与该 Note.ID 脱离关系
		return tx.Update(tag)
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

	// 最后更新 note.Tags
	note.Tags = tags
	if err := tx.Update(&note); err != nil {
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
	length := len(histories)
	if length > historyLimit {
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

// GetByTag returns notes without contents.
func (db *DB) GetByTag(name string) (notes []Note, err error) {
	tag, err := db.GetTag(name)
	if err != nil {
		return
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
	tag, err := db.GetTag(oldName)
	if err != nil {
		return err
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
	for _, noteID := range tag.NoteIDs {
		var note Note
		if err := tx.One("ID", noteID, &note); err != nil {
			return err
		}
		note.RenameTag(tag.Name, newName)
		if err := tx.UpdateField(&note, "Tags", note.Tags); err != nil {
			return err
		}
	}
	return tx.UpdateField(&tag, "Name", newName)
}
