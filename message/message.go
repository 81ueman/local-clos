package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/81ueman/local-clos/message/header"
	"github.com/81ueman/local-clos/message/keepalive"
	"github.com/81ueman/local-clos/message/open"
)

type Message interface {
	Marshal() ([]byte, error)
	UnMarshal(io.Reader) error
}

const HEADER_SIZE uint16 = 19

var ErrNotBGPMessage error = errors.New("not a BGP message")

var (
	MsgOpen      uint8 = 1
	MsgKeepalive uint8 = 4
)

func Type(m Message) (uint8, error) {
	switch m.(type) {
	case *open.Open:
		return MsgOpen, nil
	case *keepalive.Keepalive:
		return MsgKeepalive, nil
	default:
		return 0, ErrNotBGPMessage
	}
}

func Marshal(m Message) ([]byte, error) {
	ty, err := Type(m)
	if err != nil {
		return nil, fmt.Errorf("failed to get type: %v", err)
	}
	body, err := m.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %v", err)
	}
	header := header.New(HEADER_SIZE+uint16(len(body)), ty)
	head, err := header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(head, body...), nil
}

func UnMarshal(r io.Reader) (Message, error) {
	var header header.Header
	err := header.Unmarshal(r)
	// Header Validation is not implemented now (just lazy...)
	if err != nil {
		return nil, err
	}
	switch header.Type {
	case 1:
		var open open.Open
		err = open.UnMarshal(r)
		if err != nil {
			return nil, err
		}
		return &open, nil
	case 4:
		var keepalive keepalive.Keepalive
		err = keepalive.UnMarshal(r)
		if err != nil {
			return nil, err
		}
		return &keepalive, nil
	default:
		return nil, ErrNotBGPMessage
	}
}

func Send_message(w io.Writer, m Message) error {
	b, err := Marshal(m)
	if err != nil {
		return err
	}
	// write bytes to conn
	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		return err
	}
	return nil
}
