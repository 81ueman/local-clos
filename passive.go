package main

import (
	"log"
	"net"
)

func handle_connection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		for i := 0; i < n; i++ {
			log.Printf("%x", buf[i])
		}
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
