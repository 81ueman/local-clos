package update

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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

var (
	ErrInvalidPrefix = errors.New("invalid prefix")
)

func New(WithdrawnRoutes []netip.Prefix, PathAttrs []PathAttr, NLR []netip.Prefix) *UpdateMessage {
	WithdrawnRoutesLength := len(WithdrawnRoutes)
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
		rlen := (prefix.Bits()-1)/8 + 1
		err := binary.Write(buf, binary.BigEndian, uint8(prefix.Bits()))
		if err != nil {
			return nil, fmt.Errorf("failed to write a prefix length: %w", err)
		}
		pref_byte, err := prefix.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal prefix: %w", err)
		}
		buf.Write(pref_byte[:rlen-1])
		if err != nil {
			return nil, fmt.Errorf("failed to write prefix: %w", err)
		}
	}

	binary.Write(buf, binary.BigEndian, u.TotalPathAttribute)
	for _, attr := range u.PathAttrs {
		err := binary.Write(buf, binary.BigEndian, attr.AttrFlags)
		if err != nil {
			return nil, fmt.Errorf("failed to write attribute flags: %w", err)
		}
		err = binary.Write(buf, binary.BigEndian, attr.AttrTypeCode)
		if err != nil {
			return nil, fmt.Errorf("failed to write attribute type code: %w", err)
		}
		if IsExtendedLength(attr.AttrFlags) {
			err = binary.Write(buf, binary.BigEndian, uint16(len(attr.AttrValue)))
		} else {
			err = binary.Write(buf, binary.BigEndian, uint8(len(attr.AttrValue)))
		}
		if err != nil {
			return nil, fmt.Errorf("failed to write attribute length: %w", err)
		}
		err = binary.Write(buf, binary.BigEndian, attr.AttrValue)
		if err != nil {
			return nil, fmt.Errorf("failed to write attribute value: %w", err)
		}

	}
	for _, prefix := range u.NetworkLayerReachable {
		err := binary.Write(buf, binary.BigEndian, uint8(prefix.Bits()))
		if err != nil {
			return nil, fmt.Errorf("failed to write a prefix length: %w", err)
		}
		pref_byte, err := prefix.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal prefix: %w", err)
		}
		rlen := (prefix.Bits()-1)/8 + 1
		buf.Write(pref_byte[:rlen])
		if err != nil {
			return nil, fmt.Errorf("failed to write prefix: %w", err)
		}
	}
	return buf.Bytes(), nil
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
