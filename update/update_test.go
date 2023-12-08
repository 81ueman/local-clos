package update_test

import (
	"net/netip"
	"reflect"
	"testing"

	"github.com/81ueman/local-clos/update"
)

func TestNew(t *testing.T) {
	type args struct {
		WithdrawnRoutes []netip.Prefix
		PathAttrs       []update.Attrer
		NLR             []netip.Prefix
	}
	// rewrite it in another text file
	NEXT_HOP := update.NEXT_HOP(netip.AddrFrom4([4]byte{192, 168, 0, 1}))
	tests := []struct {
		name string
		args args
		want *update.UpdateMessage
	}{
		{
			name: "test1",
			args: args{
				WithdrawnRoutes: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
				PathAttrs: []update.Attrer{
					update.IGP,
					update.ASPathSegment{
						PathSegmentType: update.AS_SEQUENCE,
						AS:              []update.AS{1},
					},
					NEXT_HOP,
					update.LOCAL_PREF(100),
				},
				NLR: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
			},
			want: &update.UpdateMessage{
				WithdrawnRoutesLength: 1,
				WithdrawnRoutes: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
				TotalPathAttribute: 25,
				PathAttrs: []update.PathAttr{
					{
						AttrFlags:    update.TRANSITIVE,
						AttrTypeCode: 1,
						AttrValue:    []byte{0},
					},
					{
						AttrFlags:    update.TRANSITIVE,
						AttrTypeCode: 2,
						AttrValue: []byte{
							0x02, 0x01, 0x00, 0x01,
						},
					},
					{
						AttrFlags:    update.TRANSITIVE,
						AttrTypeCode: 3,
						AttrValue: []byte{
							0xc0, 0xa8, 0x00, 0x01,
						},
					},
					{
						AttrFlags:    update.TRANSITIVE,
						AttrTypeCode: 5,
						AttrValue: []byte{
							0x00, 0x00, 0x00, 0x64,
						},
					},
				},
				NetworkLayerReachable: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := update.New(test.args.WithdrawnRoutes, test.args.PathAttrs, test.args.NLR); reflect.DeepEqual(got, test.want) != true {
				t.Errorf("New() = %v\n want %v", got, test.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	NEXT_HOP := update.NEXT_HOP(netip.AddrFrom4([4]byte{192, 168, 0, 1}))
	tests := []struct {
		name string
		o    *update.UpdateMessage
		want []byte
	}{
		{
			name: "test1",
			o: &update.UpdateMessage{
				WithdrawnRoutesLength: 1,
				WithdrawnRoutes: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
				TotalPathAttribute: 24,
				PathAttrs: []update.PathAttr{
					update.IGP.ToAttr(),
					update.ASPathSegment{
						PathSegmentType: update.AS_SEQUENCE,
						AS:              []update.AS{1},
					}.ToAttr(),
					NEXT_HOP.ToAttr(),
					update.LOCAL_PREF(100).ToAttr(),
				},
				NetworkLayerReachable: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
			},
			want: []byte{
				0x00, 0x01, // WithdrawnRoutesLength
				24, 192, 168, 0, // WithdrawnRoutes
				0, 24,
				0x40, 0x01, 0x01, 0x00, // ORIGIN
				0x40, 0x02, 0x04, 0x02, 0x01, 0x00, 0x01, // AS_PATH
				0x40, 0x03, 0x04, 0xc0, 0xa8, 0x00, 0x01, // NEXT_HOP
				0x40, 0x05, 0x04, 0x00, 0x00, 0x00, 0x64, // LOCAL_PREF
				24, 192, 168, 0, // NetworkLayerReachable
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.o.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
			}
			if reflect.DeepEqual(got, test.want) != true {
				t.Errorf("Marshal() = %v\n want %v", got, test.want)
			}
		})
	}

}

func TestUnMarshal(t *testing.T) {
	NEXT_HOP := update.NEXT_HOP(netip.AddrFrom4([4]byte{192, 168, 0, 1}))
	tests := []struct {
		name  string
		in    []byte
		want  *update.UpdateMessage
		isErr bool
	}{
		{
			name: "test1",
			in: []byte{
				0x00, 0x01, // WithdrawnRoutesLength
				24, 192, 168, 0, // WithdrawnRoutes
				0, 24,
				0x40, 0x01, 0x01, 0x00, // ORIGIN
				0x40, 0x02, 0x04, 0x02, 0x01, 0x00, 0x01, // AS_PATH
				0x40, 0x03, 0x04, 0xc0, 0xa8, 0x00, 0x01, // NEXT_HOP
				0x40, 0x05, 0x04, 0x00, 0x00, 0x00, 0x64, // LOCAL_PREF
				24, 192, 168, 0, // NetworkLayerReachable
			},
			want: &update.UpdateMessage{
				WithdrawnRoutesLength: 1,
				WithdrawnRoutes: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
				TotalPathAttribute: 24,
				PathAttrs: []update.PathAttr{
					update.IGP.ToAttr(),
					update.ASPathSegment{
						PathSegmentType: update.AS_SEQUENCE,
						AS:              []update.AS{1},
					}.ToAttr(),
					NEXT_HOP.ToAttr(),
					update.LOCAL_PREF(100).ToAttr(),
				},
				NetworkLayerReachable: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
			},
			isErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := update.UnMarshal(test.in)
			if test.isErr && err == nil {
				t.Errorf("UnMarshal() should return error but doesn't ")
			}
			if !test.isErr && err != nil {
				t.Errorf("UnMarshal() should not return error but does: %v", err)
			}
			if reflect.DeepEqual(got, test.want) != true {
				t.Errorf("UnMarshal() = %v\n want %v", got, test.want)
			}
		})
	}

}
