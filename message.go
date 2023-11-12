package main

import (
	"github.com/81ueman/local-clos/header"
)

type Message interface {
	Marshal() ([]byte, error)
}

const header_size uint16 = 19

func Marshal(m Message, ty uint8) ([]byte, error) {
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
