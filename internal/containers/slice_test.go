package containers

import (
	"reflect"
	"testing"
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Set(tt.args.offset, tt.args.data); got != tt.want {
				t.Errorf("Slice.Set() = %v, want %v", got, tt.want)
			}
		})
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Get(tt.args.offset, tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Slice.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
