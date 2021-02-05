package database

import (
	"database/sql"
	"errors"

	"github.com/ahui2016/uglynotes/model"
	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/stmt"
	"github.com/ahui2016/uglynotes/util"
)

// 用来保存数据库的当前状态.
const (
	metadataBucket = "metadata-bucket"
	currentIdKey   = "current-id-key"
	totalSizeKey   = "total-size-key"
)

func (db *DB) GetTotalSize() (size int, err error) {
	return getTotalSize(db.DB)
}

func getCurrentID(tx TX) (id IncreaseID, err error) {
	var strID string
	row := tx.QueryRow(stmt.GetTextValue, currentIdKey)
	if err = row.Scan(&strID); err != nil {
		return
	}
	return model.ParseID(strID)
}
func initFirstID(tx TX) (err error) {
	_, err = getCurrentID(tx)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(
			stmt.InsertTextValue, currentIdKey, model.FirstID().String())
	}
	return
}
func setCurrentID(tx TX, id string) (err error) {
	_, err = tx.Exec(stmt.UpdateTextValue, id, currentIdKey)
	return
}
func getTotalSize(tx TX) (size int, err error) {
	row := tx.QueryRow(stmt.GetIntValue, totalSizeKey)
	err = row.Scan(&size)
	return
}
func initTotalSize(tx TX) (err error) {
	_, err = getTotalSize(tx)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(stmt.InsertIntValue, totalSizeKey, 0)
	}
	return
}
func increaseTotalSize(tx TX, addition int) error {
	size, err := getTotalSize(tx)
	if err != nil {
		return err
	}
	totalSize := size + addition
	if totalSize > settings.Config.DatabaseCapacity {
		return errors.New("超过数据库总容量上限")
	}
	_, err = tx.Exec(stmt.UpdateIntValue, totalSize, totalSizeKey)
	return err
}

func (db *DB) getNextID() (nextID string, err error) {
	db.Lock()
	defer db.Unlock()

	var currentID IncreaseID
	if currentID, err = getCurrentID(db.DB); err != nil {
		return
	}
	nextID = currentID.Increase().String()
	_, err = db.DB.Exec(stmt.UpdateTextValue, nextID, currentIdKey)
	return
}

func (db *DB) mustGetNextID() string {
	nextID, err := db.getNextID()
	util.Panic(err)
	return nextID
}
