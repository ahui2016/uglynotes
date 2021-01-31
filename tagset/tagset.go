package tagset

import "sort"

type Set struct {
	Map map[string]string // map[id][name]
}

type Tag struct {
	ID   string
	Name string
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
		if set.Has(k) {
			tags = append(tags, Tag{k, v})
		}
	}
	return
}

func UniqueSort(tags []Tag) (result []Tag) {
	if len(tags) == 0 {
		return
	}
	result = NewSet(tags).Slice()
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return
}
