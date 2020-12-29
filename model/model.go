package model

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/ahui2016/uglynotes/stringset"
	"github.com/ahui2016/uglynotes/util"
)

// ISO8601 需要根据服务器的具体时区来设定正确的时区
const ISO8601 = "2006-01-02T15:04:05.999+00:00"

// TitleLimit 限制标题的长度。
const TitleLimit = 50

// SizeLimit 限制每篇笔记的体积。
const SizeLimit = 1 << 19 // 512 KB

// NoteType 是一个枚举类型，用来区分 Note 的类型。
type NoteType string

const (
	Plaintext NoteType = "Plaintext"
	Markdown  NoteType = "Markdown"
)

// NewNoteType .
func NewNoteType(noteType string) NoteType {
	if strings.ToLower(noteType) == "markdown" {
		return Markdown
	}
	return Plaintext
}

// Note 表示一个数据表。
type Note struct {
	ID        string // primary key
	Type      NoteType
	Title     string
	Contents  string
	Size      int
	Tags      []string // []Tag.Name
	History   []string // []History.ID
	CreatedAt string   `storm:"index"` // ISO8601
	UpdatedAt string   `storm:"index"`
	DeletedAt string   `storm:"index"`
}

// NewNote .
func NewNote(id string, noteType NoteType) *Note {
	now := TimeNow()
	return &Note{
		ID:        id,
		Type:      noteType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetContents 在填充内容的同时设置 size, 并根据笔记类型设置标题。
// 请总是使用 SetContents 而不要直接操作 note.Contents, 以确保体积和标题正确。
func (note *Note) SetContents(contents string) error {
	title := firstLineLimit(contents, TitleLimit)
	if note.Type == Markdown {
		if mdTitle := GetMarkdownTitle(title); mdTitle != "" {
			title = mdTitle
		}
	}
	note.Title = title
	note.Contents = contents
	note.Size = len(contents)
	if note.Size > SizeLimit {
		return errors.New("size limit exceeded")
	}
	return nil
}

// SetTags 可以对标签进行除重。
// 当不需要除重时可以直接操作 note.Tags
func (note *Note) SetTags(tags []string) {
	note.Tags = stringset.Unique(tags)
}

// History 数据表，用于保存笔记的历史记录。
type History struct {
	ID        string // primary key, random
	NoteID    string `storm:"index"`
	Contents  string
	Size      int
	CreatedAt string `storm:"index"` // ISO8601
}

// Tag .
type Tag struct {
	Name      string `storm:"id"`
	NoteIDs   []string
	CreatedAt string `storm:"index"` // ISO8601
}

// NewTag .
func NewTag(name, noteID string) *Tag {
	return &Tag{
		Name:      name,
		NoteIDs:   []string{noteID},
		CreatedAt: TimeNow(),
	}
}

// Add .
func (tag *Tag) Add(noteID string) {
	if util.HasString(tag.NoteIDs, noteID) {
		return
	}
	tag.NoteIDs = append(tag.NoteIDs, noteID)
}

// TimeNow .
func TimeNow() string {
	return time.Now().Format(ISO8601)
}

// firstLineLimit 返回第一行，并限定长度，其中 s 必须事先 TrimSpace.
func firstLineLimit(s string, limit int) string {
	s += "\n"
	i := strings.IndexRune(s, '\n')
	s = s[:i]
	if len(s) > limit {
		s = s[:limit]
	}
	return s
}

func GetMarkdownTitle(s string) string {
	reTitle := regexp.MustCompile(`(^#{1,6}|>|1.|-|\*) (.+)`)
	matches := reTitle.FindStringSubmatch(s)
	// 这个 matches 要么为空，要么包含 3 个元素
	if len(matches) >= 3 {
		return matches[2]
	}
	return ""
}
