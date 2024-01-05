package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/81ueman/local-clos/message/header"
	"github.com/81ueman/local-clos/message/keepalive"
	notifiacation "github.com/81ueman/local-clos/message/notification"
	"github.com/81ueman/local-clos/message/open"
	"github.com/81ueman/local-clos/message/update"
)

type Message interface {
	Marshal() ([]byte, error)
	UnMarshal(r io.Reader, len uint16) error
}

var _ Message = &open.Open{}
var _ Message = &keepalive.Keepalive{}
var _ Message = &header.Header{}
var _ Message = &update.Update{}

const HEADER_SIZE uint16 = 19

var ErrNotBGPMessage error = errors.New("not a BGP message")

var (
	MsgTypeOpen         uint8 = 1
	MsgTypeUpdate       uint8 = 2
	MsgTypeNotification uint8 = 3
	MsgTypeKeepalive    uint8 = 4
)

func Type(m Message) (uint8, error) {
	switch m.(type) {
	case *open.Open:
		return MsgTypeOpen, nil
	case *update.Update:
		return MsgTypeUpdate, nil
	case *notifiacation.Notification:
		return MsgTypeNotification, nil
	case *keepalive.Keepalive:
		return MsgTypeKeepalive, nil
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
	err := header.UnMarshal(r, HEADER_SIZE)
	// Header Validation is not implemented now (just lazy...)
	if err != nil {
		return nil, err
	}
	l := header.Length
	switch header.Type {
	case MsgTypeOpen:
		var open open.Open
		err = open.UnMarshal(r, l)
		if err != nil {
			return nil, err
		}
		return &open, nil
	case MsgTypeUpdate:
		var update update.Update
		err = update.UnMarshal(r, l)
		if err != nil {
			return nil, err
		}
		return &update, nil
	case MsgTypeNotification:
		var notification notifiacation.Notification
		err = notification.UnMarshal(r, l)
		if err != nil {
			return nil, err
		}
		return &notification, nil
	case MsgTypeKeepalive:
		var keepalive keepalive.Keepalive
		err = keepalive.UnMarshal(r, l)
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
