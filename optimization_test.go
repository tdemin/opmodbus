package modbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_optimizeRead(t *testing.T) {
	type args struct {
		r []readOp
	}
	tests := []struct {
		name string
		args args
		want []readOp
	}{
		{
			"optimizes two requests",
			args{[]readOp{
				{2, 2},
				{4, 2},
				{7, 1},
			}},
			[]readOp{
				{2, 4},
				{7, 1},
			},
		},
		{
			"optimizes multiple requests after each other",
			args{[]readOp{
				{2, 2},
				{4, 2},
				{6, 1},
				{7, 1},
				{9, 3},
			}},
			[]readOp{
				{2, 6},
				{9, 3},
			},
		},
		{
			"skips optimization on quantity limit",
			args{[]readOp{
				{2, 4},
				{6, 2045},
			}},
			[]readOp{
				{2, 4},
				{6, 2045},
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, optimizeRead(tt.args.r), tt.name)
	}
}

func mb(a ...byte) []byte {
	// shortcuts []byte{...}
	return a
}

func Test_optimizeWrite(t *testing.T) {
	type args struct {
		w []writeOp
	}
	tests := []struct {
		name string
		args args
		want []writeOp
	}{
		{
			"optimizes two requests",
			args{[]writeOp{
				{2, 2, mb(3, 3, 4, 4)},
				{4, 2, mb(2, 3, 2, 4)},
			}},
			[]writeOp{
				{2, 4, mb(3, 3, 4, 4, 2, 3, 2, 4)},
			},
		},
		{
			"optimizes multiple requests after each other",
			args{[]writeOp{
				{2, 1, mb(3, 2)},
				{3, 1, mb(3, 4)},
				{4, 1, mb(5, 2)},
				{6, 1, mb(8, 0)},
			}},
			[]writeOp{
				{2, 3, mb(3, 2, 3, 4, 5, 2)},
				{6, 1, mb(8, 0)},
			},
		},
		{
			"doesn't care about nil data",
			args{[]writeOp{
				{2, 1, nil},
				{3, 1, mb(3, 4)},
			}},
			[]writeOp{{2, 2, mb(3, 4)}},
		},
		{
			"skips optimization on quantity limit",
			args{[]writeOp{
				{2, 10, nil},
				{12, 115, nil},
			}},
			[]writeOp{
				{2, 10, nil},
				{12, 115, nil},
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, optimizeWrite(tt.args.w), tt.name)
	}
}
