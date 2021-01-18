package model

import (
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/stringset"
	"github.com/ahui2016/uglynotes/util"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var config = settings.Config

type Patch = diffmatchpatch.Patch

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
	Deleted   bool
	CreatedAt string `storm:"index"` // ISO8601
	UpdatedAt string `storm:"index"`
}

// Note 表示一个数据表。
type Note2 struct {
	ID        string // primary key
	Type      NoteType
	Title     string
	Contents  string
	Patches   []string
	Size      int
	Tags      []string // []Tag.Name
	Deleted   bool
	CreatedAt string `storm:"index"` // ISO8601
	UpdatedAt string `storm:"index"`
}

// NewNote .
func NewNote2(id string, noteType NoteType) *Note2 {
	now := TimeNow()
	return &Note2{
		ID:        id,
		Type:      noteType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetContents 用于第一次填充内容，同时设置 size, 并根据笔记类型设置标题。
// 请总是使用 SetContents 而不要直接操作 note.Contents, 以确保体积和标题正确。
// 每篇笔记只使用一次 SetContents, 之后应使用 AddPatch.
func (note *Note2) SetContents(contents string) error {
	size, title, err := getSizeTitle(contents, note.Type)
	if err != nil {
		return err
	}
	note.Title = title
	note.Contents = contents
	note.Size = size
	return nil
}

// AddPatchNow combines AddPatch and UpdatedAtNow.
func (note *Note2) AddPatchNow(patch string) error {
	if err := note.AddPatch(patch); err != nil {
		return err
	}
	note.UpdatedAtNow()
	return nil
}

// AddPatch .
func (note *Note2) AddPatch(patch string) error {
	contents, err := patchApply(patch, note.Contents)
	if err != nil {
		return err
	}
	if err := note.SetContents(contents); err != nil {
		return err
	}
	note.Patches = append(note.Patches, patch)
	return nil
}
func getSizeTitle(contents string, noteType NoteType) (
	size int, title string, err error) {
	size = len(contents)
	if size > config.NoteSizeLimit {
		err = errors.New("size limit exceeded")
		return
	}
	title = getTitle(contents, noteType)
	if title == "" {
		err = errors.New("note title is empty")
		return
	}
	return
}
func getTitle(contents string, noteType NoteType) string {
	title := firstLineLimit(contents, config.NoteTitleLimit)
	if noteType == Markdown {
		if mdTitle := getMarkdownTitle(title); mdTitle != "" {
			title = mdTitle
		}
	}
	return title
}
func diffGNU(s string) string {
	i := strings.Index(s, "@@")
	re := regexp.MustCompile(`\\ .*`)
	return re.ReplaceAllString(s[i:], "")
}
func patchApply(patch string, text string) (string, error) {
	log.Print("patch: ", patch)
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(diffGNU(patch))
	if err != nil {
		return "", err
	}
	patched, _ := dmp.PatchApply(patches, text)
	log.Print("patched: ", patched)
	return patched, nil
}

// UpdatedAtNow updates note.UpdatedAt to TimeNow().
func (note *Note2) UpdatedAtNow() {
	note.UpdatedAt = TimeNow()
}

// SetTags 对标签进行一些验证和处理（例如除重和排序）。
// 尽量不要直接操作 note.Tags
func (note *Note2) SetTags(tags []string) error {
	sorted := stringset.UniqueSort(tags)
	if len(sorted) < 2 {
		return errors.New("too few tags (at least two)")
	}
	note.Tags = purify(sorted)
	return nil
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

// SetContentsNow combines SetContents and UpdatedAtNow.
func (note *Note) SetContentsNow(contents string) error {
	if err := note.SetContents(contents); err != nil {
		return err
	}
	note.UpdatedAtNow()
	return nil
}

// UpdatedAtNow updates note.UpdatedAt to TimeNow().
func (note *Note) UpdatedAtNow() {
	note.UpdatedAt = TimeNow()
}

// SetContents 在填充内容的同时设置 size, 并根据笔记类型设置标题。
// 请总是使用 SetContents 而不要直接操作 note.Contents, 以确保体积和标题正确。
func (note *Note) SetContents(contents string) error {
	title := firstLineLimit(contents, config.NoteTitleLimit)
	if note.Type == Markdown {
		if mdTitle := getMarkdownTitle(title); mdTitle != "" {
			title = mdTitle
		}
	}
	if title == "" {
		return errors.New("note title is empty")
	}
	note.Title = title
	note.Contents = contents
	note.Size = len(contents)
	if note.Size > config.NoteSizeLimit {
		return errors.New("size limit exceeded")
	}
	return nil
}

// SetTags 对标签进行一些验证和处理（例如除重和排序）。
// 尽量不要直接操作 note.Tags
func (note *Note) SetTags(tags []string) error {
	sorted := stringset.UniqueSort(tags)
	if len(sorted) < 2 {
		return errors.New("too few tags (at least two)")
	}
	note.Tags = purify(sorted)
	return nil
}

func purify(tags []string) (purified []string) {
	re := regexp.MustCompile(`[#;,，'"/\+\n]`)
	for i := range tags {
		purified = append(purified, re.ReplaceAllString(tags[i], ""))
	}
	return
}

// RenameTag .
func (note *Note) RenameTag(oldName, newName string) {
	note.Tags = stringset.AddAndDelete(note.Tags, oldName, newName)
}

// DeleteTag .
func (note *Note) DeleteTag(tag string) {
	i := util.StringIndex(note.Tags, tag)
	if i < 0 {
		return
	}
	note.Tags = util.DeleteFromSlice(note.Tags, i)
}

// History 数据表，用于保存笔记的历史记录。
type History struct {
	ID        string // primary key, random
	NoteID    string `storm:"index"`
	Contents  string
	Size      int
	Protected bool
	CreatedAt string `storm:"index"` // ISO8601
}

// NewHistory .
func NewHistory(contents, noteID string) *History {
	return &History{
		ID:        RandomID(),
		NoteID:    noteID,
		Contents:  contents,
		Size:      len(contents),
		CreatedAt: TimeNow(),
	}
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

// Remove .
func (tag *Tag) Remove(id string) {
	i := util.StringIndex(tag.NoteIDs, id)
	if i < 0 {
		return
	}
	tag.NoteIDs = util.DeleteFromSlice(tag.NoteIDs, i)
}

// TimeNow .
func TimeNow() string {
	return time.Now().Format(config.ISO8601)
}

// firstLineLimit 返回第一行，并限定长度，其中 s 必须事先 TrimSpace 并确保不是空字串。
// 该函数会尽量确保最后一个字符是有效的 utf8 字符，但当第一行中的全部字符都无效时，
// 则按原样返回每一行。
func firstLineLimit(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit]
	}
	s += "\n"
	i := strings.Index(s, "\n")
	firstLine := s[:i]
	for len(firstLine) > 0 {
		if utf8.ValidString(firstLine) {
			break
		}
		firstLine = firstLine[:len(firstLine)-1]
	}
	if firstLine == "" {
		firstLine = s[:i]
	}
	return firstLine
}

func getMarkdownTitle(s string) string {
	reTitle := regexp.MustCompile(`(^#{1,6}|>|1.|-|\*) (.+)`)
	matches := reTitle.FindStringSubmatch(s)
	// 这个 matches 要么为空，要么包含 3 个元素
	if len(matches) >= 3 {
		return matches[2]
	}
	return ""
}

// TagGroup 标签组，其中 Tags 应该除重和排序。
type TagGroup struct {
	ID        string   // primary key, random
	Tags      []string `storm:"unique"`
	Protected bool
	CreatedAt string `storm:"index"` // ISO8601
	UpdatedAt string `storm:"index"`
}

// NewTagGroup .
func NewTagGroup(tags []string) *TagGroup {
	now := TimeNow()
	return &TagGroup{
		ID:        RandomID(),
		Tags:      stringset.UniqueSort(tags),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// RenameTag .
func (group *TagGroup) RenameTag(oldName, newName string) {
	if util.StringIndex(group.Tags, oldName) < 0 {
		return
	}
	group.Tags = stringset.AddAndDelete(group.Tags, oldName, newName)
}
