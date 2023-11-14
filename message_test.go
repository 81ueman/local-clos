package main

import (
	"bytes"
	"testing"

	"github.com/81ueman/local-clos/keepalive"
	"github.com/81ueman/local-clos/open"
)

type dummy struct{}

func (d *dummy) Marshal() ([]byte, error) {
	return []byte{}, nil
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
