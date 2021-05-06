package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/ahui2016/uglynotes/database"
	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/util"
)

var (
	cfgFlag      = flag.String("config", "", "run with a config file")
	dbDirFlag    = flag.String("dir", "", "database directory")
	settingsFile = "settings.json"
)

var (
	config     settings.Settings
	dataDir    string // 数据库文件夹
	dbPath     string // 数据库文件
	exportPath string // 数据库导出文件
)

var (
	db          = new(database.DB)
	passwordTry = 0
)

func init() {
	flag.Parse()
	if *cfgFlag != "" {
		settingsFile = *cfgFlag
	}
	if *dbDirFlag != "" {
		dataDir = *dbDirFlag
	}

	setConfig()
	setPaths()
	util.MustMkdir(dataDir)

	// open the db here, close the db in main().
	log.Print("database:", dbPath)
	err := db.Open(dbPath)
	util.Panic(err)
	log.Print(dbPath)
}

func setPaths() {
	if dataDir == "" {
		if config.DataFolderName == "" {
			log.Fatal("config.DataFolderName is empty")
		}
		dataDir = filepath.Join(util.UserHomeDir(), config.DataFolderName)
	}
	dbPath = filepath.Join(dataDir, config.DatabaseFileName)
	exportPath = filepath.Join(dataDir, config.ExportFileName)
}

func setConfig() {
	configJSON, err := ioutil.ReadFile(settingsFile)
	config = settings.Config

	// 找不到文件或内容为空
	if err != nil || len(configJSON) == 0 {
		configJSON, err := json.MarshalIndent(config, "", "    ")
		util.Panic(err)
		util.Panic(ioutil.WriteFile(settingsFile, configJSON, 0600))
		return
	}

	// settingsFile 有内容
	util.Panic(json.Unmarshal(configJSON, &config))

	// 为了与前端 day.js 输出的格式保持一致
	if config.ISO8601 == "2006-01-02T15:04:05.999+00:00" {
		config.ISO8601 = "2006-01-02T15:04:05.999Z"
	}
}
