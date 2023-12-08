package update

import (
	"log"
	"net/netip"
)

type AttrFlags uint8
type TypeCode uint8

const (
	OPTIONAL        AttrFlags = 0x80
	TRANSITIVE      AttrFlags = 0x40
	PARTIAL         AttrFlags = 0x20
	EXTENDED_LENGTH AttrFlags = 0x10
)

type PathAttr struct {
	AttrFlags    AttrFlags
	AttrTypeCode TypeCode
	AttrValue    []byte
}
type Attrer interface {
	ToAttr() PathAttr
}

type ORIGINTYPE uint8

const (
	IGP        ORIGINTYPE = 0
	EGP        ORIGINTYPE = 1
	INCOMPLETE ORIGINTYPE = 2
)

func (o ORIGINTYPE) ToAttr() PathAttr {
	var attr PathAttr
	attr.AttrFlags = TRANSITIVE
	attr.AttrTypeCode = 1
	attr.AttrValue = []byte{byte(o)}
	return attr
}

type PathSegmentType uint8

const (
	AS_SET      PathSegmentType = 1
	AS_SEQUENCE PathSegmentType = 2
)

type PathSegmentLength uint8
type AS uint16

type ASPathSegment struct {
	PathSegmentType PathSegmentType
	AS              []AS
}

func (as ASPathSegment) ToAttr() PathAttr {
	var attr PathAttr
	attr.AttrFlags = TRANSITIVE
	attr.AttrTypeCode = 2
	attr.AttrValue = []byte{byte(as.PathSegmentType), byte(len(as.AS))}
	for _, asn := range as.AS {
		attr.AttrValue = append(attr.AttrValue, byte(asn>>8), byte(asn))
	}
	return attr
}

type NEXT_HOP netip.Addr

func (nh NEXT_HOP) ToAttr() PathAttr {
	var attr PathAttr
	attr.AttrFlags = TRANSITIVE
	attr.AttrTypeCode = 3
	bin, err := netip.Addr(nh).MarshalBinary()
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}
	attr.AttrValue = bin
	return attr
}

type LOCAL_PREF uint32

func (lp LOCAL_PREF) ToAttr() PathAttr {
	var attr PathAttr
	attr.AttrFlags = TRANSITIVE
	attr.AttrTypeCode = 5
	attr.AttrValue = []byte{byte(lp >> 24), byte(lp >> 16), byte(lp >> 8), byte(lp)}
	return attr
}
