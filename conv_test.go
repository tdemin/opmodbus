package modbus

import (
	"reflect"
	"testing"
)

func Test_uint16ToBits(t *testing.T) {
	type args struct {
		val uint16
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"converts 32",
			args{32},
			[]byte{0, 0b100000},
		},
		{
			"converts 891",
			args{891},
			[]byte{0b11, 0b1111011},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uint16ToBits(tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uint16ToBits() = %v, want %v", got, tt.want)
			}
		})
	}
}
