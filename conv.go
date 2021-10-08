package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

func uint16ToBits(val uint16) []byte {
	r := make([]byte, 2)
	binary.BigEndian.PutUint16(r, val)
	return r
}

func uint16FromBits(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

// TODO: add different float32 bit shuffle variants

func float32ToBits(val float32) []byte {
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, math.Float32bits(val))
	fp := make([]byte, 4)
	copy(fp[0:2], r[2:4])
	copy(fp[2:4], r[0:2])
	return fp
}

func float32FromBits(b []byte) float32 {
	fp := make([]byte, 4)
	copy(fp[0:2], b[2:4])
	copy(fp[2:4], b[0:2])
	return math.Float32frombits(binary.BigEndian.Uint32(fp))
}

func valueToBits(v interface{}) []byte {
	switch typeName := reflect.TypeOf(v).Name(); typeName {
	case "uint16":
		return uint16ToBits(v.(uint16))
	case "float32":
		return float32ToBits(v.(float32))
	default:
		panic("invalid value type: " + typeName)
	}
}

func valueFromBits(v []byte, t Type) interface{} {
	switch t {
	case Uint16:
		return uint16FromBits(v)
	case Float32:
		return float32FromBits(v)
	}
	panic("type " + fmt.Sprint(t) + " not implemented")
}

type Type uint

const (
	Uint16 Type = iota
	Float32
)

func (t Type) Size() uint16 {
	switch t {
	case Float32:
		return 2
	}
	return 1
}
