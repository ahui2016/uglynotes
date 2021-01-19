package settings

type Settings struct {
	DataFolderName   string
	DatabaseFileName string
	ExportFileName   string
	PasswordMaxTry   int
	Password         string
	Address          string

	// MaxAge for session
	// 有效单位是 "s", "m", "h"
	MaxAge string

	// NoteTitleLimit 限制标题的长度。
	// 注意：该限制还需要在 public/util.js 中设置（为了做前端限制）
	NoteTitleLimit int

	// NoteSizeLimit 限制每篇笔记的体积。
	// 注意：该限制还需要在 public/util.js 中设置（为了做前端限制）
	NoteSizeLimit int

	// MaxBodySize 单个文件上限, 通常应设置为等于 NoteSizeLimit
	MaxBodySize int

	// DatabaseCapacity 整个数据库的体积上限
	DatabaseCapacity int

	// ISO8601 需要根据服务器的具体时区来设定正确的时区
	// 比如，如果是北京时间，则应设为 "2006-01-02T15:04:05.999+08:00"
	ISO8601 string

	// HistoryLimit 限制每篇笔记可保留的历史版本数量上限。
	// 该参数已经废除，现在历史版本系统升级了，不设上限。
	HistoryLimit int

	// TagGroupLimit 限制标签组数量上限。
	// 当超过上限时，不受保护的标签组会被覆盖。可通过点击 "protect" 按钮保护标签。
	TagGroupLimit int
}

var Config = Default()

func Default() Settings {
	return Settings{
		DataFolderName:   "uglynotes_data_folder",
		DatabaseFileName: "uglynotes.db",
		ExportFileName:   "uglynotes.json",
		PasswordMaxTry:   100,
		Password:         "abc",
		Address:          "127.0.0.1:80",
		MaxAge:           "2400h", // 24 * 100 = 100 days
		NoteTitleLimit:   200,
		NoteSizeLimit:    1 << 19, // 512 KB
		MaxBodySize:      1 << 19,
		DatabaseCapacity: 1 << 20 * 10, // 10MB
		ISO8601:          "2006-01-02T15:04:05.999+00:00",
		TagGroupLimit:    100,
	}
}
