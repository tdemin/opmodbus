package containers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice_Set(t *testing.T) {
	type args struct {
		offset int
		data   []byte
	}
	tests := []struct {
		name string
		s    Slice
		args args
		want int
	}{
		{
			"normally",
			NewSlice(4),
			args{2, []byte{3, 4}},
			2,
		},
		{
			"with overlap",
			NewSlice(4),
			args{2, []byte{3, 4, 5}},
			2,
		},
		{
			"to empty slice",
			NewSlice(0),
			args{3, []byte{4, 5}},
			0,
		},
		{
			"with negative offset",
			NewSlice(3),
			args{-1, []byte{4, 5}},
			0,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.s.Set(tt.args.offset, tt.args.data), tt.name)
	}
}

func TestSlice_Get(t *testing.T) {
	type args struct {
		offset int
		size   int
	}
	tests := []struct {
		name string
		s    Slice
		args args
		want []byte
	}{
		{
			"normally",
			Slice([]byte{3, 4, 5, 6}),
			args{1, 2},
			[]byte{4, 5},
		},
		{
			"with overlap",
			Slice([]byte{3, 4, 5}),
			args{1, 3},
			[]byte{4, 5},
		},
		{
			"with offset overlap",
			Slice([]byte{3, 4, 5}),
			args{3, 3},
			nil,
		},
		{
			"with negative size",
			Slice([]byte{3, 4}),
			args{1, -1},
			nil,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.s.Get(tt.args.offset, tt.args.size), tt.name)
	}
}
