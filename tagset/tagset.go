package tagset

import (
	"sort"

	"github.com/ahui2016/uglynotes/util"
)

type Set struct {
	Map map[string]string // map[id][name]
}

type Tag struct {
	ID   string
	Name string
}

func TagsFrom(tagNames []string) (tags []Tag) {
	for i := range tagNames {
		tag := Tag{Name: tagNames[i]}
		tags = append(tags, tag)
	}
	return
}

func ToNames(tags []Tag) (names []string) {
	for i := range tags {
		names = append(names, tags[i].Name)
	}
	return
}

func NewSet(tags []Tag) *Set {
	set := &Set{make(map[string]string)}
	for _, tag := range tags {
		set.Map[tag.ID] = tag.Name
	}
	return set
}

func (set *Set) Has(tag Tag) (ok bool) {
	_, ok = set.Map[tag.ID]
	return
}

func (set *Set) Add(tag Tag) {
	set.Map[tag.ID] = tag.Name
}

func (set *Set) Delete(tag Tag) {
	delete(set.Map, tag.ID)
}

func (set *Set) Slice() (tags []Tag) {
	for k, v := range set.Map {
		tags = append(tags, Tag{k, v})
	}
	return
}

func SortByID(tags []Tag) []Tag {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].ID < tags[j].ID
	})
	return tags
}

func SortByIDToString(tags []Tag) string {
	var sorted []string
	for _, tag := range SortByID(tags) {
		sorted = append(sorted, tag.ID)
	}
	blob := util.MustMarshal(sorted)
	return string(blob)
}

func UniqueSortName(tags []Tag) (result []Tag) {
	if len(tags) == 0 {
		return
	}
	result = NewSet(tags).Slice()
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return
}
