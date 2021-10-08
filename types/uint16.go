package types

import (
	"encoding/binary"
	"fmt"
)

// Uint16 is a regular unsigned big endian int that fits in a single
// Modbus register.
type Uint16 uint16

func (u Uint16) Bytes() []byte {
	r := make([]byte, 2)
	binary.BigEndian.PutUint16(r, uint16(u))
	return r
}

func (u Uint16) Size() uint16 {
	return 1
}

func (Uint16) Converter() Converter {
	return func(b []byte) (Value, error) {
		if l := len(b); l != 2 {
			return nil, fmt.Errorf("%w: bytes of size %v", ErrInvalidInput, l)
		}

		return Uint16(binary.BigEndian.Uint16(b)), nil
	}
}

// Uint16Type is provided for use as Type.
const Uint16Type = Uint16(0)
