package message

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/81ueman/local-clos/message/keepalive"
	"github.com/81ueman/local-clos/message/open"
)

func TestType(t *testing.T) {
	type args struct {
		m Message
	}
	tests := []struct {
		name    string
		args    args
		want    uint8
		wantErr bool
	}{
		{
			name: "open",
			args: args{
				m: &open.Open{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "keepalive",
			args: args{
				m: &keepalive.Keepalive{},
			},
			want:    4,
			wantErr: false,
		},
		{
			name: "dummy_fails",
			args: args{
				m: &dummy{},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := Type(tt.args.m)
			if got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

type dummy struct{}

func (d *dummy) Marshal() ([]byte, error) {
	return []byte{}, nil
}

func (d *dummy) UnMarshal(r io.Reader) error {
	return nil
}

func TestMarshal(t *testing.T) {

	type args struct {
		m Message
	}
	marker := make([]byte, 16)
	for i := 0; i < 16; i++ {
		marker[i] = 0xff
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "open",
			args: args{
				m: &open.Open{
					Version:  4,
					AS:       1,
					Holdtime: 1,
					Id:       1,
				},
			},
			want:    append(marker, []byte{0, 28, 1, 4, 0, 1, 0, 1, 0, 0, 0, 1}...),
			wantErr: false,
		},
		{
			name: "keepalive",
			args: args{
				m: &keepalive.Keepalive{},
			},
			want:    append(marker, []byte{0, 19, 4}...),
			wantErr: false,
		},
		{
			name: "dummy_fails",
			args: args{
				m: &dummy{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnMarshal(t *testing.T) {
	marker := make([]byte, 16)
	for i := 0; i < 16; i++ {
		marker[i] = 0xff
	}

	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "open",
			args: args{
				r: bytes.NewReader(
					append(marker, []byte{0, 28, 1, 4, 0, 1, 0, 1, 0, 0, 0, 1}...),
				),
			},
			want: &open.Open{
				Version:  4,
				AS:       1,
				Holdtime: 1,
				Id:       1,
			},
			wantErr: false,
		},
		{
			name: "keepalive",
			args: args{
				bytes.NewReader(
					append(marker, []byte{0, 19, 4}...),
				),
			},
			want:    &keepalive.Keepalive{},
			wantErr: false,
		},
		{
			name: "dummy_fails",
			args: args{
				r: bytes.NewReader([]byte{255, 255, 255, 255, 255, 255, 255, 255, 0, 19, 4}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnMarshal(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnMarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) != true {
				t.Errorf("UnMarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}
