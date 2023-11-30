package update

import (
	"bytes"
	"encoding/binary"
	"net/netip"
)

type AttrFlags uint8
type TypeCode uint8

const (
	ORIGIN           TypeCode = 1
	AS_PATH          TypeCode = 2
	NEXT_HOP         TypeCode = 3
	MULTI_EXIT_DISC  TypeCode = 4
	LOCAL_PREF       TypeCode = 5
	ATOMIC_AGGREGATE TypeCode = 6
	AGGREGATOR       TypeCode = 7
)

type PathAttr struct {
	AttrFlags    AttrFlags
	AttrTypeCode TypeCode
	AttrValue    []byte
}

type UpdateMessage struct {
	WithdrawnRoutesLength uint16
	WithdrawnRoutes       []netip.Prefix
	TotalPathAttribute    uint16
	PathAttrs             []PathAttr
	NetworkLayerReachable []netip.Prefix
}

func New(WithdrawnRoutes []netip.Prefix, PathAttrs []PathAttr, NLR []netip.Prefix) *UpdateMessage {
	WithdrawnRoutesLength := 0
	for _, prefix := range WithdrawnRoutes {
		WithdrawnRoutesLength += 1
		WithdrawnRoutesLength += prefix.Bits() / 8
	}
	TotalPathAttribute := 0
	for _, attr := range PathAttrs {
		TotalPathAttribute += 1 // AttrFlags
		TotalPathAttribute += 1 // AttrTypeCode
		if IsExtendedLength(attr.AttrFlags) {
			TotalPathAttribute += 2 // AttrLength
		} else {
			TotalPathAttribute += 1 // AttrLength
		}
		TotalPathAttribute += len(attr.AttrValue) // AttrValue
	}

	return &UpdateMessage{
		WithdrawnRoutesLength: uint16(WithdrawnRoutesLength),
		WithdrawnRoutes:       WithdrawnRoutes,
		TotalPathAttribute:    uint16(TotalPathAttribute),
		PathAttrs:             PathAttrs,
		NetworkLayerReachable: NLR,
	}
}

func (u *UpdateMessage) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, u.WithdrawnRoutesLength)
	if err != nil {
		return nil, err
	}
	for _, prefix := range u.WithdrawnRoutes {
		err = binary.Write(buf, binary.BigEndian, prefix)
		if err != nil {
			return nil, err
		}
	}
}

func IsOptional(flag AttrFlags) bool {
	return flag&0x80 == 0x80
}

func IsTransitive(flag AttrFlags) bool {
	return flag&0x40 == 0x40
}

func IsPartial(flag AttrFlags) bool {
	return flag&0x20 == 0x20
}

func IsExtendedLength(flag AttrFlags) bool {
	return flag&0x10 == 0x10
}
