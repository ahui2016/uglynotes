package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/ahui2016/uglynotes/database"
	"github.com/ahui2016/uglynotes/util"
)

const (
	dataFolderName   = "uglynotes_data_folder"
	databaseFileName = "uglynotes.db"
	passwordMaxTry   = 5
	defaultPassword  = "abc"
	defaultAddress   = "127.0.0.1:80"

	// 99 days, for session
	maxAge = 99 * time.Hour * 24

	// 单个文件上限
	maxBodySize = 1024 * 32 // 32 KB

	// 整个数据库上限
	databaseCapacity = 1 << 20 // 1MB
)

var (
	db          = new(database.DB)
	passwordTry = 0
)

var (
	dataDir = filepath.Join(util.UserHomeDir(), dataFolderName)
	dbPath  = filepath.Join(dataDir, databaseFileName)
)

func init() {
	util.MustMkdir(dataDir)

	// open the db here, close the db in main().
	err := db.Open(maxAge, databaseCapacity, dbPath)
	util.Panic(err)
	log.Print(dbPath)
}
