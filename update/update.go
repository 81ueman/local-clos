package update

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net/netip"

	"github.com/81ueman/local-clos/header"
)

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

func New(WithdrawnRoutes []netip.Prefix, PathAttrs []Attrer, NLR []netip.Prefix) *UpdateMessage {
	WithdrawnRoutesLength := len(WithdrawnRoutes)
	TotalPathAttribute := 0
	for _, attr := range PathAttrs {
		TotalPathAttribute += 1 // AttrFlags
		TotalPathAttribute += 1 // AttrTypeCode
		if IsExtendedLength(attr.ToAttr().AttrFlags) {
			TotalPathAttribute += 2 // AttrLength
		} else {
			TotalPathAttribute += 1 // AttrLength
		}
		TotalPathAttribute += len(attr.ToAttr().AttrValue) // AttrValue
	}

	attrs := make([]PathAttr, 0)
	for _, attr := range PathAttrs {
		attrs = append(attrs, attr.ToAttr())
	}

	return &UpdateMessage{
		WithdrawnRoutesLength: uint16(WithdrawnRoutesLength),
		WithdrawnRoutes:       WithdrawnRoutes,
		TotalPathAttribute:    uint16(TotalPathAttribute),
		PathAttrs:             attrs,
		NetworkLayerReachable: NLR,
	}
}

// ひどいのでリファクタ必須
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
		buf.Write(pref_byte[:rlen])
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

func UnMarshal(b []byte) (*UpdateMessage, error) {
	withdrawnRoutesLength := binary.BigEndian.Uint16(b[0:2])
	b = b[2:]
	withdrawnRoutes := make([]netip.Prefix, 0)
	for i := 0; i < int(withdrawnRoutesLength); i++ {
		prefixLength := b[0]
		b = b[1:]

		octet := (prefixLength-1)/8 + 1
		prefix := netip.Prefix{}
		prefix_bin := make([]byte, 5)
		copy(prefix_bin, b[:octet])
		b = b[octet:]
		prefix_bin[4] = prefixLength
		err := prefix.UnmarshalBinary(prefix_bin)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal prefix: %w", err)
		}
		withdrawnRoutes = append(withdrawnRoutes, prefix)
	}

	totalPathAttrLength := binary.BigEndian.Uint16(b[0:2])
	b = b[2:]
	pathAttrs := make([]PathAttr, 0)
	for i := 0; i < int(totalPathAttrLength); {
		attrFlags := AttrFlags(b[0])
		b = b[1:]
		i += 1
		attrTypeCode := TypeCode(b[0])
		b = b[1:]
		i += 1
		var attrLength uint16
		if IsExtendedLength(attrFlags) {
			attrLength = binary.BigEndian.Uint16(b[0:2])
			b = b[2:]
			i += 2
		} else {
			attrLength = uint16(b[0])
			b = b[1:]
			i += 1
		}
		attrValue := b[:attrLength]
		b = b[attrLength:]
		pathAttrs = append(pathAttrs, PathAttr{
			AttrFlags:    attrFlags,
			AttrTypeCode: attrTypeCode,
			AttrValue:    attrValue,
		})
		i += int(attrLength)
	}
	NLR := make([]netip.Prefix, 0)
	for len(b) > 0 {
		prefixLength := b[0]
		octet := (prefixLength-1)/8 + 1
		b = b[1:]
		prefix := netip.Prefix{}

		prefix_bin := make([]byte, 5)
		copy(prefix_bin, b[:octet])
		prefix_bin[4] = prefixLength
		err := prefix.UnmarshalBinary(prefix_bin)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal prefix: %w", err)
		}
		b = b[octet:]
		NLR = append(NLR, prefix)
	}
	return &UpdateMessage{
		WithdrawnRoutesLength: withdrawnRoutesLength,
		WithdrawnRoutes:       withdrawnRoutes,
		TotalPathAttribute:    totalPathAttrLength,
		PathAttrs:             pathAttrs,
		NetworkLayerReachable: NLR,
	}, nil
}

func Read(r io.Reader) (*UpdateMessage, error) {
	b := new(bytes.Buffer)
	if _, err := io.CopyN(b, r, 19); err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	h := header.Header{}
	binary.Read(b, binary.BigEndian, &h)

	b = new(bytes.Buffer)

	log.Println("length: ", h.Length)
	if _, err := io.CopyN(b, r, int64(h.Length)-19); err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	msg, err := UnMarshal(b.Bytes())
	if err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
	}
	return msg, nil

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
