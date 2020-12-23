package model

import "time"

// ISO8601 需要根据服务器的具体时区来设定正确的时区
const ISO8601 = "2006-01-02T15:04:05.999+00:00"

// NoteType 是一个枚举类型，用来区分 Note 的类型。
type NoteType string

const (
	Plaintext NoteType = "Plaintext"
	Markdown  NoteType = "Markdown"
)

// Note 表示一个数据表。
type Note struct {
	ID        string // primary key
	Type      NoteType
	Name      string
	Contents  string
	Size      int64
	Tags      []string // []Tag.Name
	History   []string // []History.ID
	CreatedAt string   `storm:"index"` // ISO8601
	UpdatedAt string   `storm:"index"`
	DeletedAt string   `storm:"index"`
}

func NewNote(id string, noteType NoteType) *Note {
	now := TimeNow()
	return &Note{
		ID:        id,
		Type:      noteType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// History 数据表，用于保存笔记的历史记录。
type History struct {
	ID        string // primary key, random
	NoteID    string `storm:"index"`
	Contents  string
	Size      int64
	CreatedAt string `storm:"index"` // ISO8601
}

// Tag .
type Tag struct {
	Name      string `storm:"id"`
	NoteIDs   []string
	CreatedAt string `storm:"index"` // ISO8601
}

// TimeNow .
func TimeNow() string {
	return time.Now().Format(ISO8601)
}
