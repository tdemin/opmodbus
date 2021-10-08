package memory

type Map struct {
	buf []byte
}

func NewMap(size int) *Map {
	return &Map{
		buf: make([]byte, size),
	}
}

func (a *Map) Set(offset int, data []byte) int {
	return copy(a.buf[offset:], data)
}

func (a *Map) Get(offset, size int) []byte {
	return a.buf[offset : offset+size]
}
