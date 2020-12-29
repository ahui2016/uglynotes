package database

import (
	"errors"
	"sync"
	"time"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/util"
	"github.com/asdine/storm/v3"
	"github.com/gofiber/fiber/v2/middleware/session"
)

const cookieName = "uglynotesCookie"
const maxHistory = 10

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
	err3 := db.initTotalSize()
	return util.WrapErrors(err1, err2, err3)
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
func (db *DB) GetByID(id string) (*Note, error) {
	var note Note
	err := db.DB.One("ID", id, &note)
	return &note, err
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

// AllNotes fetches all notes, sorted by "UpdatedAt".
func (db *DB) AllNotes() (notes []Note, err error) {
	err = db.DB.AllByIndex("UpdatedAt", &notes)
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
	return db.DB.Save(note)
}
