package main

import (
	"log"
	"net"
	"time"

	"github.com/81ueman/local-clos/header"
)

func active_mode() {
	laddr := &net.TCPAddr{
		IP:   net.ParseIP("192.168.0.1"),
		Port: 179,
	}
	raddr := &net.TCPAddr{
		IP:   net.ParseIP("192.168.0.2"),
		Port: 179,
	}
	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	header := header.New(0xffff, 0xff)
	for {
		hbytes, err := header.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		conn.Write(hbytes)
		time.Sleep(1 * time.Second)
	}
}
