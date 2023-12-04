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
		PathAttrs       []update.PathAttr
		NLR             []netip.Prefix
	}
	// rewrite it in another text file
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
				PathAttrs: []update.PathAttr{
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.ORIGIN,
						AttrValue:    []byte{0x00},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.AS_PATH,
						AttrValue:    []byte{0x02, 0x01, 0x00},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.NEXT_HOP,
						AttrValue:    []byte{0xc0, 0xa8, 0x00, 0x01},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.LOCAL_PREF,
						AttrValue:    []byte{0x00, 0x00, 0x00, 0x64},
					},
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
				TotalPathAttribute: 24,
				PathAttrs: []update.PathAttr{
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.ORIGIN,
						AttrValue:    []byte{0x00},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.AS_PATH,
						AttrValue:    []byte{0x02, 0x01, 0x00},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.NEXT_HOP,
						AttrValue:    []byte{0xc0, 0xa8, 0x00, 0x01},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.LOCAL_PREF,
						AttrValue:    []byte{0x00, 0x00, 0x00, 0x64},
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
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.ORIGIN,
						AttrValue:    []byte{0x00},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.AS_PATH,
						AttrValue:    []byte{0x02, 0x01, 0x00},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.NEXT_HOP,
						AttrValue:    []byte{0xc0, 0xa8, 0x00, 0x01},
					},
					{
						AttrFlags:    0x40,
						AttrTypeCode: update.LOCAL_PREF,
						AttrValue:    []byte{0x00, 0x00, 0x00, 0x64},
					},
				},
				NetworkLayerReachable: []netip.Prefix{
					netip.MustParsePrefix("192.168.0.0/24"),
				},
			},
			want: []byte{
				0x00, 0x01, // WithdrawnRoutesLength
				24, 192, 168, 0, // WithdrawnRoutes
				24,
				0x40, 0x01, 0x01, 0x00, // ORIGIN
				0x40, 0x02, 0x03, 0x02, 0x01, 0x00, // AS_PATH
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
