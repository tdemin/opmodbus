package types

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Float32CDAB is a 32-bit IEEE floating point value where the 4 bytes
// order is swapped from ABCD to CDAB before transmission.
type Float32CDAB float32

func (f Float32CDAB) Bytes() []byte {
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, math.Float32bits(float32(f)))
	fp := make([]byte, 4)
	copy(fp[0:2], r[2:4])
	copy(fp[2:4], r[0:2])
	return fp
}

func (f Float32CDAB) Size() uint16 {
	return 2
}

func (Float32CDAB) Converter() Converter {
	return func(b []byte) (Value, error) {
		if l := len(b); l != 4 {
			return nil, fmt.Errorf("%w: bytes of size %v", ErrInvalidInput, l)
		}

		fp := make([]byte, 4)
		copy(fp[0:2], b[2:4])
		copy(fp[2:4], b[0:2])
		return Float32CDAB(math.Float32frombits(binary.BigEndian.Uint32(fp))), nil
	}
}

// Float32CDABType is provided for use as Type.
const Float32CDABType = Float32CDAB(0)

// Float32CDAB is a 32-bit IEEE floating point value with the regular
// byte order.
type Float32 float32

func (f Float32) Bytes() []byte {
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, math.Float32bits(float32(f)))
	return r
}

func (f Float32) Size() uint16 {
	return 2
}

func (Float32) Converter() Converter {
	return func(b []byte) (Value, error) {
		if l := len(b); l != 4 {
			return nil, fmt.Errorf("%w: bytes of size %v", ErrInvalidInput, l)
		}

		return Float32(math.Float32frombits(binary.BigEndian.Uint32(b))), nil
	}
}

// Float32Type is provided for use as Type.
const Float32Type = Float32(0)
