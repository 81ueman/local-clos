// Description: This package is used to create a BGP header for the packet
package header

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Header struct {
	Marker [16]byte
	Length uint16
	Type   uint8
}

func New(length uint16, Type uint8) *Header {
	marker := make([]byte, 16)
	for i := 0; i < 16; i++ {
		marker[i] = 0xff
	}
	return &Header{
		Marker: [16]byte(marker),
		Length: length,
		Type:   Type,
	}
}

// Marshal is used to convert the header into a byte array
func (h *Header) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, h)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *Header) UnMarshal(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, h)
	if err != nil {
		return err
	}
	return nil
}
