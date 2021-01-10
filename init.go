package main

import (
	"log"
	"path/filepath"

	"github.com/ahui2016/uglynotes/database"
	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/util"
)

var (
	db          = new(database.DB)
	passwordTry = 0
)

var (
	dataDir = filepath.Join(util.UserHomeDir(), settings.DataFolderName)
	dbPath  = filepath.Join(dataDir, settings.DatabaseFileName)
)

func init() {
	util.MustMkdir(dataDir)

	// open the db here, close the db in main().
	err := db.Open(dbPath)
	util.Panic(err)
	log.Print(dbPath)
}
