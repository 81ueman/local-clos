package header

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		length uint16
		Type   uint8
	}
	marker := [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	tests := []struct {
		name string
		args args
		want *Header
	}{
		{
			name: "test1",
			args: args{
				length: 1,
				Type:   1,
			},
			want: &Header{
				Marker: marker,
				Length: 1,
				Type:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.length, tt.args.Type); *got != *tt.want {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {

	marker := [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	tests := []struct {
		name string
		h    *Header
		want []byte
	}{
		{
			name: "bigendian_1",
			h: &Header{
				Marker: marker,
				Length: 1,
				Type:   1,
			},
			want: []byte{255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 0, 1, 1},
		},
		{
			name: "bigendian_2",
			h: &Header{
				Marker: marker,
				Length: 256,
				Type:   1,
			},
			want: []byte{255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 1, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.h.Marshal(); !bytes.Equal(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}

}
