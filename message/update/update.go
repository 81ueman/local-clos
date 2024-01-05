package update

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/netip"
)

type AttrFlags uint8

var (
	AttrFlagsOptional       AttrFlags = 0x80
	AttrFlagsTransitive     AttrFlags = 0x40
	AttrFlagsPartial        AttrFlags = 0x20
	AttrFlagsExtendedLength AttrFlags = 0x10
)

func (a *AttrFlags) Optional() bool {
	return (*a & AttrFlagsOptional) != 0
}

func (a *AttrFlags) Transitive() bool {
	return (*a & AttrFlagsTransitive) != 0
}

func (a *AttrFlags) Partial() bool {
	return (*a & AttrFlagsPartial) != 0
}

func (a *AttrFlags) ExtendedLength() bool {
	return (*a & AttrFlagsExtendedLength) != 0
}

type AttrType uint8

var (
	AttrTypeOrigin          AttrType = 1
	AttrTypeASPath          AttrType = 2
	AttrTypeNextHop         AttrType = 3
	AttrTypeMultiExitDisc   AttrType = 4
	AttrTypeLocalPref       AttrType = 5
	AttrTypeAtomicAggregate AttrType = 6
	AttrTypeAggregator      AttrType = 7
)

type Origin uint8

var (
	OriginIGP Origin = 0
	OriginEGP Origin = 1
	OriginINC Origin = 2
)

func (o *Origin) marshal() ([]byte, error) {
	b := make([]byte, 4)
	b[0] = byte(AttrFlagsTransitive)
	b[1] = byte(AttrTypeOrigin)
	b[2] = 1
	b[3] = byte(*o)
	return b, nil
}

type VALUE_SEGMENT_TYPE uint8

var (
	VALUE_SEGMENT_AS_SET      VALUE_SEGMENT_TYPE = 1
	VALUE_SEGMENT_AS_SEQUENCE VALUE_SEGMENT_TYPE = 2
)

type AS_PATH struct {
	VALUE_SEGMENT VALUE_SEGMENT_TYPE
	AS_SEQUENCE   []uint16
}

func (a *AS_PATH) marshal() ([]byte, error) {
	b := make([]byte, 3+1+1+2*len(a.AS_SEQUENCE))
	b[0] = byte(AttrFlagsTransitive)
	b[1] = byte(AttrTypeASPath)
	b[2] = uint8(1 + 1 + 2*len(a.AS_SEQUENCE))
	b[3] = byte(a.VALUE_SEGMENT)
	b[4] = uint8(len(a.AS_SEQUENCE))
	for i, as := range a.AS_SEQUENCE {
		binary.BigEndian.PutUint16(b[5+2*i:], as)
	}
	return b, nil
}

type NEXT_HOP netip.Addr

func (n *NEXT_HOP) marshal() ([]byte, error) {
	b := make([]byte, 1+1+1+4)
	b[0] = byte(AttrFlagsTransitive)
	b[1] = byte(AttrTypeNextHop)
	b[2] = 4
	bin, err := netip.Addr(*n).MarshalBinary()
	if err != nil {
		return nil, err
	}
	if len(bin) != 4 {
		return nil, fmt.Errorf("invalid next hop address: %v", *n)
	}
	copy(b[3:], bin)
	return b, nil
}

type LOCAL_PREF uint32

func (l *LOCAL_PREF) marshal() ([]byte, error) {
	b := make([]byte, 1+1+1+4)
	b[0] = byte(AttrFlagsTransitive)
	b[1] = byte(AttrTypeLocalPref)
	b[2] = 4
	binary.BigEndian.PutUint32(b[3:], uint32(*l))
	return b, nil
}

type ATOMIC_AGGREGATE bool

type Update struct {
	WithdrawnRoutes                     []netip.Prefix
	PathAttrOrigin                      Origin
	PathAttrASPath                      AS_PATH
	PathAttrNextHop                     NEXT_HOP
	PathAttrLocalPref                   LOCAL_PREF
	NetworkLayerReachabilityInformation []netip.Prefix
}

func prefixToBytes(prefix netip.Prefix) ([]byte, error) {
	pLen := (prefix.Bits()-1)/8 + 1

	b := make([]byte, 1+pLen)
	b[0] = byte(prefix.Bits())

	bin, err := prefix.Masked().Addr().MarshalBinary()
	if err != nil {
		return nil, err
	}

	bin = bin[:pLen]
	copy(b[1:], bin)
	return b, nil

}

// binだとsliceのコピーが発生して遅いかもしれないが実装の簡略化を優先
// bytes.Bufferを使ったほうがパフォーマンスは良いかもしれない
func (u *Update) Marshal() ([]byte, error) {
	var bin []byte
	bin = binary.BigEndian.AppendUint16(bin, uint16(len(u.WithdrawnRoutes)))
	for _, prefix := range u.WithdrawnRoutes {
		b, err := prefixToBytes(prefix)
		if err != nil {
			return nil, err
		}
		bin = append(bin, b...)
	}

	originBin, err := u.PathAttrOrigin.marshal()
	if err != nil {
		return nil, err
	}
	aspathBin, err := u.PathAttrASPath.marshal()
	if err != nil {
		return nil, err
	}
	nexthopBin, err := u.PathAttrNextHop.marshal()
	if err != nil {
		return nil, err
	}
	localprefBin, err := u.PathAttrLocalPref.marshal()
	if err != nil {
		return nil, err
	}
	TotalPathAttrLen := len(originBin) + len(aspathBin) + len(nexthopBin) + len(localprefBin)
	bin = binary.BigEndian.AppendUint16(bin, uint16(TotalPathAttrLen))
	bin = append(bin, originBin...)
	bin = append(bin, aspathBin...)
	bin = append(bin, nexthopBin...)
	bin = append(bin, localprefBin...)
	for _, prefix := range u.NetworkLayerReachabilityInformation {
		b, err := prefixToBytes(prefix)
		if err != nil {
			return nil, err
		}
		bin = append(bin, b...)
	}
	return bin, nil
}

func (u *Update) UnMarshal(r io.Reader, length uint16) error {
	var withdrawnLength uint16
	var err error
	if err = binary.Read(r, binary.BigEndian, &withdrawnLength); err != nil {
		return err
	}
	log.Printf("withdrawnLength: %v", withdrawnLength)
	withdrawnRoutesBin := make([]byte, withdrawnLength)
	io.ReadFull(r, withdrawnRoutesBin)
	for i := 0; i < int(withdrawnLength); {
		prefixBin := make([]byte, 5)
		prefixLen := withdrawnRoutesBin[i]
		copy(prefixBin, withdrawnRoutesBin[i+1:i+1+(int(prefixLen)-1)/8+1])
		prefixBin[4] = prefixLen
		var prefix netip.Prefix
		err := prefix.UnmarshalBinary(prefixBin)
		if err != nil {
			return err
		}
		u.WithdrawnRoutes = append(u.WithdrawnRoutes, prefix)
		i += 1 + len(prefixBin)
	}
	log.Printf("withdrawnRoutes: %v", u.WithdrawnRoutes)

	var pathAttrLen uint16
	err = binary.Read(r, binary.BigEndian, &pathAttrLen)
	if err != nil {
		return err
	}
	log.Printf("pathAttrLen: %v", pathAttrLen)
	pathAttrBin := make([]byte, pathAttrLen)
	io.ReadFull(r, pathAttrBin)
	for i := 0; i < int(pathAttrLen); {
		attrflags := AttrFlags(pathAttrBin[i])
		i += 1
		attrType := AttrType(pathAttrBin[i])
		i += 1
		var attrLen uint
		if attrflags.ExtendedLength() {
			attrLen = uint(binary.BigEndian.Uint16(pathAttrBin[i:]))
			i += 2
		} else {
			attrLen = uint(pathAttrBin[i])
			i += 1
		}
		log.Printf("attrflags: %v, attrType: %v, attrLen: %v", attrflags, attrType, attrLen)

		switch attrType {
		case AttrTypeOrigin:
			u.PathAttrOrigin = Origin(pathAttrBin[i])
			i += 1
			if attrLen != 1 {
				return fmt.Errorf("invalid origin length: %v", attrLen)
			}
		case AttrTypeASPath:
			u.PathAttrASPath.VALUE_SEGMENT = VALUE_SEGMENT_TYPE(pathAttrBin[i])
			i += 1
			segmentLen := pathAttrBin[i]
			i += 1
			for j := 0; j < int(segmentLen); j++ {
				as := binary.BigEndian.Uint16(pathAttrBin[i+j*2:])
				u.PathAttrASPath.AS_SEQUENCE = append(u.PathAttrASPath.AS_SEQUENCE, as)
			}
			i += int(segmentLen) * 2
			if attrLen != uint(1+1+2*segmentLen) {
				return fmt.Errorf("invalid aspath length: %v", attrLen)
			}
		case AttrTypeNextHop:
			var nexthop netip.Addr
			err := nexthop.UnmarshalBinary(pathAttrBin[i : i+4])
			i += 4
			if err != nil {
				return err
			}
			u.PathAttrNextHop = NEXT_HOP(nexthop)
			if attrLen != 4 {
				return fmt.Errorf("invalid nexthop length: %v", attrLen)
			}
		case AttrTypeLocalPref:
			u.PathAttrLocalPref = LOCAL_PREF(binary.BigEndian.Uint32(pathAttrBin[i:]))
			i += 4
			if attrLen != 4 {
				return fmt.Errorf("invalid localpref length: %v", attrLen)
			}
		}
	}
	NLRlength := length - 2 - withdrawnLength - 2 - pathAttrLen
	log.Printf("NLRlength: %v", NLRlength)
	// 最初からbufioでいいじゃん
	rd := bufio.NewReader(r)
	for i := 0; i < int(NLRlength); {
		plen, err := rd.ReadByte()
		log.Printf("plen: %v", plen)
		if err != nil {
			return err
		}
		i += 1

		prefixBin := make([]byte, 5)
		io.ReadFull(rd, prefixBin[:(plen-1)/8+1])
		i += int((plen-1)/8 + 1)
		prefixBin[4] = plen
		var prefix netip.Prefix
		if err = prefix.UnmarshalBinary(prefixBin); err != nil {
			return err
		}
		u.NetworkLayerReachabilityInformation = append(u.NetworkLayerReachabilityInformation, prefix)
	}
	return nil
}
