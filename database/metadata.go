package database

import (
	"errors"

	"github.com/ahui2016/uglynotes/model"
	"github.com/asdine/storm/v3"
)

// 用来保存数据库的当前状态.
const (
	metadataBucket = "metadata-bucket"
	currentIdKey   = "current-id-key"
	totalSizeKey   = "total-size-key"
)

func (db *DB) initFirstID() (err error) {
	_, err = db.getCurrentID()
	if err != nil && err != storm.ErrNotFound {
		return
	}
	if err == storm.ErrNotFound {
		id := model.FirstID()
		return db.DB.Set(metadataBucket, currentIdKey, id)
	}
	return
}

func (db *DB) getCurrentID() (id IncreaseID, err error) {
	err = db.DB.Get(metadataBucket, currentIdKey, &id)
	return
}

func (db *DB) getNextID() (nextID IncreaseID, err error) {
	var currentID IncreaseID
	if currentID, err = db.getCurrentID(); err != nil {
		return
	}
	nextID = currentID.Increase()
	err = db.DB.Set(metadataBucket, currentIdKey, &nextID)
	return
}

func (db *DB) initTotalSize() (err error) {
	_, err = db.GetTotalSize()
	if err != nil && err != storm.ErrNotFound {
		return
	}
	if err == storm.ErrNotFound {
		return db.setTotalSize(0)
	}
	return
}

// GetTotalSize .
func (db *DB) GetTotalSize() (size int64, err error) {
	err = db.DB.Get(metadataBucket, totalSizeKey, &size)
	return
}

func (db *DB) setTotalSize(size int64) error {
	return db.DB.Set(metadataBucket, totalSizeKey, size)
}

func (db *DB) checkTotalSize(addition int64) error {
	totalSize, err := db.GetTotalSize()
	if err != nil {
		return err
	}
	if totalSize+addition > db.capacity {
		return errors.New("超过数据库总容量上限")
	}
	return nil
}

// increaseTotalSize 用于向数据库添加或删除单项内容时更新总体积。
// 添加时，应先使用 db.checkTotalSize, 再使用 db.Save, 最后使才使用 db.increaseTotalSize
// 删除时，应先获取即将删除项目的体积，再删除，最后使用 db.increaseTotalSize, 此时 addition 应为负数。
func (db *DB) increaseTotalSize(addition int64) error {
	totalSize, err := db.GetTotalSize()
	if err != nil {
		return err
	}
	return db.setTotalSize(totalSize + addition)
}

// recountTotalSize 用于一次性删除多个项目时重新计算数据库总体积。
func (db *DB) recountTotalSize() error {
	return nil
}
