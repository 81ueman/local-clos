package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/81ueman/local-clos/header"
	"github.com/81ueman/local-clos/keepalive"
	"github.com/81ueman/local-clos/open"
)

type Message interface {
	Marshal() ([]byte, error)
}

type open_MSG struct {
	header.Header
	open.Open
}

type keepalive_MSG struct {
	header.Header
}

const HEADER_SIZE uint16 = 19

var ErrNotBGPMessage error = errors.New("not a BGP message")

func Type(m Message) (uint8, error) {
	switch m.(type) {
	case *open.Open:
		return 1, nil
	case *keepalive.Keepalive:
		return 4, nil
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

func send_message(conn net.Conn, m Message) error {
	b, err := Marshal(m)
	if err != nil {
		return err
	}
	// write bytes to conn
	_, err = io.Copy(conn, bytes.NewReader(b))
	if err != nil {
		return err
	}
	return nil
}
