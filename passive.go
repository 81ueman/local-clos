package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/81ueman/local-clos/keepalive"
	"github.com/81ueman/local-clos/open"
)

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
				err := send_message(session.Conn, open_msg)
				if err != nil {
					log.Fatalf("failed to send message: %v", err)
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
			msg := keepalive.New()
			err := send_message(session.Conn, msg)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			var keepalive_msg keepalive_MSG
			err = binary.Read(session.Conn, binary.BigEndian, &keepalive_msg)
			if err != nil {
				log.Printf("error: %v", err)
				return
			}
			log.Printf("keepalive_msg: %v", keepalive_msg)
			session.State = Established
		case Established:
			log.Println("session established")
			select {}
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
