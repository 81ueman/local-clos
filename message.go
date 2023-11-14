package main

import (
	"errors"
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

const header_size uint16 = 19

var ErrNotBGPMessage error = errors.New("not a BGP message")

func Marshal(m Message) ([]byte, error) {
	var ty uint8
	switch m.(type) {
	case *open.Open:
		ty = 1
	case *keepalive.Keepalive:
		ty = 4
	default:
		return nil, ErrNotBGPMessage
	}
	body, err := m.Marshal()
	if err != nil {
		return nil, err
	}
	header := header.New(header_size+uint16(len(body)), ty)
	head, err := header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(head, body...), nil
}

func send_message(conn net.Conn, m Message) error {
	bytes, err := Marshal(m)
	if err != nil {
		return err
	}
	_, err = conn.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
