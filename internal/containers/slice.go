package containers

// Slice is a byte slice safe from panics on invalid indices and such.
type Slice []byte

func NewSlice(size int) Slice {
	return make(Slice, size)
}

// Set performs copy() and returns its result. It is up to the caller to
// check whether the result is equal to len(data).
func (s Slice) Set(offset int, data []byte) int {
	if len(s) <= offset || offset < 0 {
		return 0
	}
	return copy(s[offset:], data)
}

// Get returns a subslice of Slice with given offset/size. If offset is
// larger than Slice length, it will return nil. If offset+size overlaps
// the Slice, it will return Slice[offset:].
func (s Slice) Get(offset, size int) []byte {
	if offset < 0 || size < 0 {
		return nil
	}
	if len(s) <= offset {
		return nil
	}
	if len(s) <= offset+size {
		return s[offset:]
	}
	return s[offset : offset+size]
}
