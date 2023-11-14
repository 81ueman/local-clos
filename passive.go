package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/81ueman/local-clos/header"
	"github.com/81ueman/local-clos/open"
)

type open_MSG struct {
	header.Header
	open.Open
}

func handle_bgp_passive(ifi net.Interface) {
	session := Session{
		State:               Idle,
		ConnectRetryCounter: 0,
		ConnectRetryTime:    120 * time.Second,
		HoldTime:            180 * time.Second,
		KeepaliveTime:       60 * time.Second,
	}

	events := make(chan Event)

	for {
		log.Print("session state: ", session.State)
		switch session.State {
		case Idle:
			go func(session *Session) {
				conn, err := wait_tcp(ifi)
				if err != nil {
					log.Fatalf("failed to wait tcp: %v", err)
					events <- TcpConnectionFails
				}
				session.Conn = conn
				events <- Tcp_CR_Acked
			}(&session)
			session.State = Active
		case Connect:
			// Should't reach here
			log.Fatal("shouldn't reach here")
		case Active:
			event := <-events
			switch event {
			case Tcp_CR_Acked:
				open_msg := open.New(4, 65000, 180, 0)
				bytes, err := Marshal(open_msg)
				if err != nil {
					log.Fatalf("failed to marshal: %v", err)
				}
				_, err = session.Conn.Write(bytes)
				if err != nil {
					log.Fatalf("failed to write: %v", err)
				}
				session.State = OpenSent
			case TcpConnectionFails:
				session.State = Active
			default:
				log.Fatalf("unknown event: %v", event)
			}
		case OpenSent:
			msg := open_MSG{}
			err := binary.Read(session.Conn, binary.BigEndian, &msg)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			log.Printf("msg: %v", msg)
			session.State = OpenConfirm

		case OpenConfirm:
		case Established:
		}

	}
}
func passive_mode() {

	ifis, err := net.Interfaces()
	if err != nil {
		log.Fatalf("failed to get interfaces: %v", err)
	}
	for _, ifi := range ifis {
		if is_loopback(ifi) {
			continue
		}
		go handle_bgp_passive(ifi)
	}
	select {}
}
