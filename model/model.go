package model

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ahui2016/uglynotes/settings"
	"github.com/ahui2016/uglynotes/stringset"
	"github.com/ahui2016/uglynotes/tagset"
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

type OldNote struct {
	ID        string // primary key
	Type      NoteType
	Title     string
	Patches   []string
	Size      int
	Tags      []string
	Deleted   bool
	RemindAt  string `storm:"index"`
	CreatedAt string `storm:"index"` // ISO8601
	UpdatedAt string `storm:"index"`
}

type Note struct {
	ID        string // primary key
	Type      NoteType
	Title     string
	Patches   []string
	Size      int
	Tags      []tagset.Tag
	Deleted   bool
	RemindAt  string // ISO8601 
	CreatedAt string
	UpdatedAt string 
}

func NoteFrom(oldNote OldNote) Note {
	var tags []tagset.Tag
	for _, name := range oldNote.Tags {
		tags = append(tags, tagset.Tag{Name: name})
	}
	return Note{
		ID:        oldNote.ID,
		Type:      oldNote.Type,
		Title:     oldNote.Title,
		Patches:   oldNote.Patches,
		Size:      oldNote.Size,
		Tags:      tags,
		Deleted:   oldNote.Deleted,
		CreatedAt: oldNote.CreatedAt,
		UpdatedAt: oldNote.UpdatedAt,
	}
}

// NewNote .
func NewNote(id, title, patch string, noteType NoteType, tagNames []string) (
	*Note, error) {
	note := newNote(id, noteType)
	err1 := note.AddPatchSetTitle(patch, title)
	err2 := note.SetNewTags(tagNames)
	return note, util.WrapErrors(err1, err2)
}
func newNote(id string, noteType NoteType) *Note {
	now := TimeNow()
	return &Note{
		ID:        id,
		Type:      noteType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddPatchNow combines AddPatchSetTitle and UpdatedAtNow.
func (note *Note) AddPatchNow(patch, contents string) error {
	if err := note.AddPatchSetTitle(patch, contents); err != nil {
		return err
	}
	note.UpdatedAt = TimeNow()
	return nil
}

// AddPatchSetTitle .
func (note *Note) AddPatchSetTitle(patch, contents string) error {
	note.SetTitle(contents)
	return note.AddPatch(patch)
}

// AddPatch 填充内容，同时设置 size。
// 请总是使用 AddPatch 而不要直接操作 note.Patches, 以确保体积和标题正确。
func (note *Note) AddPatch(patch string) error {
	if err := note.resetSize(len(patch)); err != nil {
		return err
	}
	note.Patches = append(note.Patches, patch)
	return nil
}

func (note *Note) UpdateTitleSizeNow(title string, patchSize int) error {
	note.SetTitle(title)
	note.UpdatedAt = TimeNow()
	return note.resetSize(patchSize)
}

func (note *Note) resetSize(patchSize int) error {
	size := note.Size + patchSize
	if size > config.NoteSizeLimit {
		return errors.New("size limit exceeded")
	}
	note.Size = size
	return nil
}

// SetTitle 设置限定长度的标题，其中 contents 必须事先 TrimSpace 并确保不是空字串。
func (note *Note) SetTitle(contents string) {
	title := firstLineLimit(contents, config.NoteTitleLimit)
	if note.Type == Markdown {
		if mdTitle := getMarkdownTitle(title); mdTitle != "" {
			title = mdTitle
		}
	}
	note.Title = title
}

func diffGNU(s string) string {
	i := strings.Index(s, "@@")
	re := regexp.MustCompile(`\\ .*`)
	return re.ReplaceAllString(s[i:], "")
}

func patchApply(patch string, text string) (string, error) {
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(diffGNU(patch))
	if err != nil {
		return "", err
	}
	patched, _ := dmp.PatchApply(patches, text)
	return patched, nil
}

// SetNewTags 对标签进行一些验证和处理（例如除重和排序）。
// 尽量不要直接操作 note.Tags
func (note *Note) SetNewTags(tagNames []string) error {
	sorted := stringset.UniqueSort(tagNames)
	if len(sorted) < 2 {
		return errors.New("too few tags (at least two)")
	}
	purified := purify(sorted)
	note.Tags = tagset.TagsFrom(purified)
	return nil
}

func purify(tags []string) []string {
	re := regexp.MustCompile(`[#;,，'"/\+\n]`)
	for i := range tags {
		tags[i] = re.ReplaceAllString(tags[i], "")
	}
	return tags
}

// Tag .
type Tag struct {
	ID        string
	Name      string
	Count     int
	CreatedAt string // ISO8601
}

// NewTag .
func NewTag(name string) *Tag {
	return &Tag{
		ID:        RandomID(),
		Name:      name,
		Count:     0,
		CreatedAt: TimeNow(),
	}
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
	ID        string // primary key, random
	Tags      []string
	Protected bool
	CreatedAt string // ISO8601
	UpdatedAt string
}

// NewTagGroup .
func NewTagGroup(tags []string) *TagGroup {
	now := TimeNow()
	return &TagGroup{
		ID:        RandomID(),
		Tags:      tags,
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
