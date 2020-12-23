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
	History    = model.History
	Tag        = model.Tag
	IncreaseID = model.IncreaseID
)

// DB .
type DB struct {
	path     string
	capacity int64
	DB       *storm.DB
	Sess     *session.Store

	// 只在 package database 外部使用锁，不在 package database 内部使用锁。
	sync.Mutex
}

// Open .
func (db *DB) Open(maxAge time.Duration, cap int64, dbPath string) (err error) {
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

// Insert .
func (db *DB) Insert(note *Note) error {
	if err := db.checkTotalSize(note.Size); err != nil {
		return err
	}
	if err := db.checkExist(note.ID); err != nil {
		return err
	}
	if err := db.DB.Save(note); err != nil {
		return err
	}
	return db.increaseTotalSize(note.Size)
}

// 检查 ID 冲突
func (db *DB) checkExist(id string) error {
	_, err := db.getByID(id)
	if err == nil {
		return errors.New("id: " + id + " already exists")
	}
	return nil
}

func (db *DB) getByID(id string) (*Note, error) {
	var note Note
	err := db.DB.One("ID", id, &note)
	return &note, err
}
