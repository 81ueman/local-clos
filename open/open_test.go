package open

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		version  uint8
		AS       uint16
		holdtime uint16
		id       uint32
	}
	tests := []struct {
		name string
		args args
		want *Open
	}{
		{
			name: "test1",
			args: args{
				version:  4,
				AS:       65000,
				holdtime: 180,
				id:       1,
			},
			want: &Open{
				Version:  4,
				AS:       65000,
				Holdtime: 180,
				Id:       1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.version, tt.args.AS, tt.args.holdtime, tt.args.id); *got != *tt.want {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name string
		o    *Open
		want []byte
	}{
		{
			name: "bigendian_1",
			o: &Open{
				Version:  4,
				AS:       1,
				Holdtime: 1,
				Id:       1,
			},
			want: []byte{4, 0, 1, 0, 1, 0, 0, 0, 1},
		},
		{
			name: "bigendian_2",
			o: &Open{
				Version:  4,
				AS:       1,
				Holdtime: 256,
				Id:       256,
			},
			want: []byte{4, 0, 1, 1, 0, 0, 0, 1, 0},
		},
		{
			name: "bigendian_3",
			o: &Open{
				Version:  4,
				AS:       256,
				Holdtime: 1,
				Id:       256 * 256,
			},
			want: []byte{4, 1, 0, 0, 1, 0, 1, 0, 0},
		},
		{
			name: "bigendian_4",
			o: &Open{
				Version:  4,
				AS:       256,
				Holdtime: 256,
				Id:       256 * 256 * 256,
			},
			want: []byte{4, 1, 0, 1, 0, 1, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.o.Marshal(); !bytes.Equal(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}

}
