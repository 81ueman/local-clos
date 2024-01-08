package main

import (
	"net/netip"
	"reflect"
	"testing"

	"github.com/81ueman/local-clos/message/update"
)

func TestUpdate(t *testing.T) {
	RibAdj := RibAdj{}
	msg := update.Update{
		WithdrawnRoutes: []netip.Prefix{},
		PathAttrOrigin:  update.Origin(1),
		PathAttrASPath: update.AS_PATH{
			VALUE_SEGMENT: update.VALUE_SEGMENT_AS_SEQUENCE,
			AS_SEQUENCE:   []uint16{1, 2, 3},
		},
		PathAttrNextHop:   update.NEXT_HOP(netip.MustParseAddr("192.168.0.1")),
		PathAttrLocalPref: update.LOCAL_PREF(100),
		NetworkLayerReachabilityInformation: []netip.Prefix{
			netip.MustParsePrefix("192.168.0.0/24"),
		},
	}
	RibAdj.Update(msg)
	if len(RibAdj) != 1 {
		t.Fatal("invalid len")
	}
	for _, prefix := range msg.NetworkLayerReachabilityInformation {
		entry, ok := RibAdj[prefix]
		if !ok {
			t.Fatal("NLR is not registered")
		}
		if entry.ORIGIN != msg.PathAttrOrigin {
			t.Fatal("invalid origin")
		}
		if !reflect.DeepEqual(entry.AS_PATH, msg.PathAttrASPath) {
			t.Fatal("invalid as path")
		}
		if entry.NEXT_HOP != msg.PathAttrNextHop {
			t.Fatal("invalid next hop")
		}
		if entry.LOCAL_PREF != msg.PathAttrLocalPref {
			t.Fatal("invalid local pref")
		}

	}
}
