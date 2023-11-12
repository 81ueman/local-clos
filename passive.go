package main

import (
	"encoding/binary"
	"log"
	"net"

	"github.com/81ueman/local-clos/header"
	"github.com/81ueman/local-clos/open"
)

type open_MSG struct {
	header.Header
	open.Open
}

func handle_connection(conn net.Conn) {
	defer conn.Close()
	for {
		msg := open_MSG{}
		err := binary.Read(conn, binary.BigEndian, &msg)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		log.Printf("msg: %v", msg)
	}
}
func passive_mode() {
	l, err := net.Listen("tcp", ":179")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handle_connection(conn)
	}
}
