package stringset

// Set .
type Set struct {
	Map map[string]bool
}

// NewSet convert a string slice to a set.
func NewSet(arr []string) *Set {
	set := &Set{make(map[string]bool)}
	for _, v := range arr {
		set.Map[v] = true
	}
	return set
}

// Slice convert the set to a string slice.
func (set *Set) Slice() []string {
	var arr []string
	for key := range set.Map {
		if set.Map[key] {
			arr = append(arr, key)
		}
	}
	return arr
}

// Unique 利用 Set 对 arr 进行除重处理。
func Unique(arr []string) []string {
	return NewSet(arr).Slice()
}
