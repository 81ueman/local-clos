package update

import (
	"encoding/binary"
	"net/netip"
	"testing"
)

func TestMarshalOrigin(t *testing.T) {
	o := Origin(1)
	b, err := o.marshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 4 {
		t.Fatal("invalid length")
	}
	if b[0] != byte(AttrFlagsTransitive) {
		t.Fatal("invalid attr flags")
	}
	if b[1] != byte(AttrTypeOrigin) {
		t.Fatal("invalid attr type")
	}
	if b[2] != 1 {
		t.Fatal("invalid attr length")
	}
	if b[3] != byte(o) {
		t.Fatal("invalid origin")
	}
}

func TestMarshalAS_PATH(t *testing.T) {
	a := AS_PATH{
		VALUE_SEGMENT: VALUE_SEGMENT_AS_SEQUENCE,
		AS_SEQUENCE:   []uint16{1, 2, 3},
	}
	b, err := a.marshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 11 {
		t.Fatal("invalid length")
	}
	if b[0] != byte(AttrFlagsTransitive) {
		t.Fatal("invalid attr flags")
	}
	if b[1] != byte(AttrTypeASPath) {
		t.Fatal("invalid attr type")
	}
	if b[2] != 8 {
		t.Fatal("invalid attr length")
	}
	if b[3] != byte(VALUE_SEGMENT_AS_SEQUENCE) {
		t.Fatal("invalid value segment type")
	}
	if b[4] != 3 {
		t.Fatal("invalid number of ASes")
	}
	if binary.BigEndian.Uint16(b[5:7]) != 1 {
		t.Fatal("invalid AS number")
	}
	if binary.BigEndian.Uint16(b[7:9]) != 2 {
		t.Fatal("invalid AS number")
	}
	if binary.BigEndian.Uint16(b[9:11]) != 3 {
		t.Fatal("invalid AS number")
	}
}

func TestMarshalLOCAL_PREF(t *testing.T) {
	l := LOCAL_PREF(1)
	b, err := l.marshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 7 {
		t.Fatal("invalid length")
	}
	if b[0] != byte(AttrFlagsTransitive) {
		t.Fatal("invalid attr flags")
	}
	if b[1] != byte(AttrTypeLocalPref) {
		t.Fatal("invalid attr type")
	}
	if b[2] != 4 {
		t.Fatal("invalid attr length")
	}
	if binary.BigEndian.Uint32(b[3:7]) != 1 {
		t.Fatal("invalid local pref")
	}
}

func TestMarshalNEXT_HOP(t *testing.T) {
	n := NEXT_HOP(netip.MustParseAddr("1.2.3.4"))
	b, err := n.marshal()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 7 {
		t.Fatal("invalid length")
	}
	if b[0] != byte(AttrFlagsTransitive) {
		t.Fatal("invalid attr flags")
	}
	if b[1] != byte(AttrTypeNextHop) {
		t.Fatal("invalid attr type")
	}
	if b[2] != 4 {
		t.Fatal("invalid attr length")
	}
	if b[3] != 1 {
		t.Fatal("invalid next hop")
	}
	if b[4] != 2 {
		t.Fatal("invalid next hop")
	}
	if b[5] != 3 {
		t.Fatal("invalid next hop")
	}
	if b[6] != 4 {
		t.Fatal("invalid next hop")
	}
}

func concatSlice(slices ...[]byte) []byte {
	var b []byte
	for _, slice := range slices {
		b = append(b, slice...)
	}
	return b
}

func TestMarshalUpdate(t *testing.T) {
	origin := Origin(OriginEGP)
	as_path := AS_PATH{
		VALUE_SEGMENT: VALUE_SEGMENT_AS_SEQUENCE,
		AS_SEQUENCE:   []uint16{1, 2, 3},
	}
	next_hop := NEXT_HOP(netip.MustParseAddr("1.2.3.4"))
	local_pref := LOCAL_PREF(1)
	origin_bin, _ := origin.marshal()
	as_path_bin, _ := as_path.marshal()
	next_hop_bin, _ := next_hop.marshal()
	local_pref_bin, _ := local_pref.marshal()
	attrlen := len(origin_bin) + len(as_path_bin) + len(next_hop_bin) + len(local_pref_bin)

	tests := []struct {
		name   string
		update Update
		want   []byte
	}{
		{
			"no withdraws",
			Update{
				WithdrawnRoutes:   []netip.Prefix{},
				PathAttrOrigin:    origin,
				PathAttrASPath:    as_path,
				PathAttrNextHop:   next_hop,
				PathAttrLocalPref: local_pref,
				NetworkLayerReachabilityInformation: []netip.Prefix{
					netip.MustParsePrefix("1.2.3.0/24"),
				},
			},
			concatSlice(
				[]byte{0, 0}, // withdrawn routes length
				binary.BigEndian.AppendUint16([]byte{}, uint16(attrlen)), // total path attr length
				origin_bin,
				as_path_bin,
				next_hop_bin,
				local_pref_bin,
				[]byte{24, 1, 2, 3}, // NLRI
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.update.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if len(b) != len(tt.want) {
				t.Errorf("invalid length: \ngot %v, \nwant %v", b, tt.want)
			}
			for i := range b {
				if b[i] != tt.want[i] {
					t.Fatalf("invalid byte at position %d: got %d, want %d", i, b[i], tt.want[i])
				}
			}
		})
	}

}
