package main

import (
	"log"
	"net"
	"time"
)

func handle_bgp(ifi net.Interface) {
	session := Session{
		State:               Idle,
		ConnectRetryCounter: 0,
		ConnectRetryTime:    120 * time.Second,
		HoldTime:            180 * time.Second,
		KeepaliveTime:       60 * time.Second,
	}

	events := make(chan Event)

	for {
		log.Println("session state: ", session.State)
		switch session.State {
		case Idle:
			go func(session *Session) {
				conn, err := start_tcp(ifi)
				if err != nil {
					events <- TcpConnectionFails
				}
				session.Conn = conn
				events <- Tcp_CR_Acked
			}(&session)

			session.State = Connect
		case Connect:
			event := <-events
			switch event {
			case Tcp_CR_Acked:
				session.State = OpenSent
			case TcpConnectionFails:
				session.State = Active
			default:
				log.Fatalf("unknown event: %v", event)
			}
		case Active:
		case OpenSent:
			log.Println("session conn: ", session.Conn.RemoteAddr().String())
		case OpenConfirm:
		case Established:
		default:
			log.Fatalf("unknown state: %v", session.State)
		}
		time.Sleep(1 * time.Second)
	}
	/*bytes, err := Marshal(open.New(4, 65000, 180, 0))
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}
	_, err = conn.Write(bytes)
	if err != nil {
		log.Fatalf("failed to write: %v", err)
	}
	time.Sleep(1 * time.Second)*/
}

func active_mode() {
	ifis, err := net.Interfaces()
	if err != nil {
		log.Fatalf("failed to get interfaces: %v", err)
	}
	for _, ifi := range ifis {
		if is_loopback(ifi) {
			continue
		}
		log.Printf("sending bgp from %v", ifi.Name)
		go handle_bgp(ifi)
	}
	select {}
}
