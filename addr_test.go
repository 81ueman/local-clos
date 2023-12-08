package main

import (
	"net"
	"net/netip"
	"reflect"
	"testing"
)

func TestAddrToPrefix(t *testing.T) {
	type args struct {
		addr net.Addr
		mask int
	}
	tests := []struct {
		name string
		args args
		want netip.Prefix
	}{
		{
			name: "test1",
			args: args{
				addr: &net.IPAddr{
					IP: net.ParseIP("192.168.0.1"),
				},
				mask: 24,
			},
			want: netip.MustParsePrefix("192.168.0.0/24"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddrToPrefix(tt.args.addr, tt.args.mask); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddrToPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
