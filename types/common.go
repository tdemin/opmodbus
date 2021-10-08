// Package types provides the most common data types for use with
// Modbus.
package types

import "errors"

// Type represents a Modbus data type.
type Type interface {
	// Size returns data type size in number of Modbus registers (2
	// bytes each).
	Size() uint16
	// Converter returns a conversion function. Function is expected to
	// be non-nil.
	Converter() Converter
}

// Value represents something that can be put into Modbus packet data
// section.
type Value interface {
	// Bytes returns a non-nil, even-length byte representation of a
	// Value. If the Value also implements Type, length of the returned
	// slice is expected to be equal to Size() * 2.
	Bytes() []byte
}

// Converter is a function that converts a byte representation into
// Value. If a byte representation is invalid for this data type, an
// error of type ErrInvalidInput is returned, and Value will be nil.
type Converter func([]byte) (Value, error)

// ErrInvalidInput is returned when a byte representation cannot be
// converted into the specified type.
var ErrInvalidInput = errors.New("invalid byte input")
