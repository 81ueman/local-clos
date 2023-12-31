package open

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrInvalidLength error = errors.New("invalid length")
)

type Open struct {
	Version  uint8
	AS       uint16
	Holdtime uint16
	Id       uint32
}

func New(version uint8, AS uint16, holdtime uint16, id uint32) *Open {
	return &Open{
		Version:  version,
		AS:       AS,
		Holdtime: holdtime,
		Id:       id,
	}
}

func (o *Open) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, o)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *Open) UnMarshal(r io.Reader, l uint16) error {
	// length check is ignored now(just lazy)
	err := binary.Read(r, binary.BigEndian, o)
	if err != nil {
		return err
	}
	return nil
}
