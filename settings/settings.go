package settings

import "time"

const (
	DataFolderName   = "uglynotes_data_folder"
	DatabaseFileName = "uglynotes.db"
	ExportFileName   = "uglynotes.json"
	PasswordMaxTry   = 5
	DefaultPassword  = "abc"
	DefaultAddress   = "127.0.0.1:80"
)

// MaxAge for session, 99 days
const MaxAge = 99 * time.Hour * 24

// NoteSizeLimit 限制每篇笔记的体积。
const NoteSizeLimit = 1 << 19 // 512 KB

// MaxBodySize 单个文件上限
const MaxBodySize = NoteSizeLimit

// DatabaseCapacity 整个数据库上限
const DatabaseCapacity = 1 << 20 * 10 // 10MB

// ISO8601 需要根据服务器的具体时区来设定正确的时区
const ISO8601 = "2006-01-02T15:04:05.999+00:00"

// 比如，如果是北京时间，则应设定如下：
// const ISO8601 = "2006-01-02T15:04:05.999+08:00"

// NoteTitleLimit 限制标题的长度。
const NoteTitleLimit = 200

// HistoryLimit 限制每篇笔记可保留的历史版本数量上限。
const HistoryLimit = 100

// TagGroupLimit 限制标签组数量上限。
const TagGroupLimit = 100
