package compare

// Slice is an interface for Slice function.
type Slice interface {
	// Len is the length about this slice.
	Len() int

	// ID is used to identified data.
	ID(i int) string
}

// UniqueSlice is used to compare two slice and return indexes about
// added and deleted. Added are indexes in current slice, deleted are
// indexes in previous slice. Each data must be unique in one slice.
func UniqueSlice(previous, current Slice) (added, deleted []int) {
	prevLen := previous.Len()
	curLen := current.Len()
	// key is data ID, value is data index
	prevMap := make(map[string]int, prevLen)
	curMap := make(map[string]int, curLen)
	for i := 0; i < prevLen; i++ {
		prevMap[previous.ID(i)] = i
	}
	for i := 0; i < curLen; i++ {
		curMap[current.ID(i)] = i
	}
	// find added items
	for item, i := range curMap {
		if _, ok := prevMap[item]; !ok {
			added = append(added, i)
		}
	}
	// find deleted items
	for item, i := range prevMap {
		if _, ok := curMap[item]; !ok {
			deleted = append(deleted, i)
		}
	}
	return
}

type stringSlice []string

func (s stringSlice) Len() int {
	return len(s)
}

func (s stringSlice) ID(i int) string {
	return s[i]
}

// UniqueStrings is used to compare two string slice, Added are indexed
// in current slice, deleted are indexes in previous slice.each string
// must be unique in one string slice.
func UniqueStrings(previous, current []string) (added, deleted []int) {
	return UniqueSlice(stringSlice(previous), stringSlice(current))
}
