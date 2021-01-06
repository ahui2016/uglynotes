package stringset

import "sort"

// Set .
type Set struct {
	Map map[string]bool
}

// NewSet convert a string slice to a set.
func NewSet(arr []string) *Set {
	set := newSet()
	for _, v := range arr {
		set.Map[v] = true
	}
	return set
}

func newSet() *Set {
	return &Set{make(map[string]bool)}
}

// Has .
func (set *Set) Has(item string) bool {
	return set.Map[item]
}

// Add .
func (set *Set) Add(item string) {
	set.Map[item] = true
}

// Delete .
func (set *Set) Delete(item string) {
	set.Map[item] = false
}

// Intersect .
func (set *Set) Intersect(other *Set) *Set {
	result := newSet()
	for key := range set.Map {
		if other.Has(key) {
			result.Add(key)
		}
	}
	return result
}

// Slice convert the set to a string slice.
func (set *Set) Slice() (arr []string) {
	for key := range set.Map {
		if set.Has(key) {
			arr = append(arr, key)
		}
	}
	return
}

// UniqueSort 利用 Set 对 arr 进行除重和排序。
func UniqueSort(arr []string) (result []string) {
	if len(arr) == 0 {
		return
	}
	result = NewSet(arr).Slice()
	sort.Strings(result)
	return
}

// AddAndDelete 利用 Set 对 arr 进行添加和删除操作，返回排序结果。
// 适用于类似于重命名的情形。
func AddAndDelete(arr []string, toDelete, toAdd string) []string {
	set := NewSet(arr)
	set.Delete(toDelete)
	set.Add(toAdd)
	result := set.Slice()
	sort.Strings(result)
	return result
}

// Intersect 取 group 里全部集合的交集。
func Intersect(group []*Set) *Set {
	length := len(group)
	if length == 0 {
		return newSet()
	}
	result := group[0]
	for i := 1; i < length; i++ {
		result = result.Intersect(group[i])
	}
	return result
}
