package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"time"

	"github.com/81ueman/local-clos/header"
	"github.com/81ueman/local-clos/keepalive"
	"github.com/81ueman/local-clos/open"
	"github.com/81ueman/local-clos/update"
)

func (s *Session) Idle(ifi net.Interface, active bool) {
	go func(session *Session) {
		var conn net.Conn
		var err error
		if active {
			conn, err = start_tcp(ifi)
		} else {
			conn, err = wait_tcp(ifi)
		}
		if err != nil {
			log.Fatalf("failed to handle tcp connection: %v", err)
			s.Events <- TcpConnectionFails
		}
		session.Conn = conn
		s.Events <- Tcp_CR_Acked
	}(s)

	if active {
		s.State = Connect
	} else {
		s.State = Active
	}
}

func (s *Session) Connect() {
	event := <-s.Events
	switch event {
	case Tcp_CR_Acked:
		open_msg := open.New(4, 65000, 180, 0)
		err := send_message(s.Conn, open_msg)
		if err != nil {
			log.Fatalf("failed to write: %v", err)
		}
		s.State = OpenSent
	case TcpConnectionFails:
		s.State = Idle
	default:
		log.Fatalf("unknown event: %v", event)
	}
}

func (s *Session) Active() {
	event := <-s.Events
	switch event {
	case Tcp_CR_Acked:
		open_msg := open.New(4, 65000, 180, 0)
		err := send_message(s.Conn, open_msg)
		if err != nil {
			log.Fatalf("failed to send message: %v", err)
		}
		s.State = OpenSent
	case TcpConnectionFails:
		s.State = Active
	default:
		log.Fatalf("unknown event: %v", event)
	}
}

func (s *Session) OpenSent() {
	log.Println("session conn: ", s.Conn.RemoteAddr().String())
	msg := open_MSG{}
	err := binary.Read(s.Conn, binary.BigEndian, &msg)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("msg: %v", msg)
	s.State = OpenConfirm
}

func (s *Session) OpenConfirm() {
	keepalive := keepalive.New()
	err := send_message(s.Conn, keepalive)
	if err != nil {
		log.Fatalf("failed to write: %v", err)
	}
	var msg keepalive_MSG
	if err = binary.Read(s.Conn, binary.BigEndian, &msg); err != nil {
		log.Printf("error: %v", err)
		return
	}
	log.Printf("msg: %v", msg)
	s.State = Established
}

func (s *Session) Established() {

	sample := update.New(
		[]netip.Prefix{},
		[]update.PathAttr{
			{
				AttrFlags:    0x40,
				AttrTypeCode: update.ORIGIN,
				AttrValue:    []byte{0x00},
			},
			{
				AttrFlags:    0x40,
				AttrTypeCode: update.AS_PATH,
				AttrValue:    []byte{byte(update.AS_SEQUENCE), 0x01, 0x10, 0x00},
			},
			{
				AttrFlags:    0x40,
				AttrTypeCode: update.NEXT_HOP,
				AttrValue:    []byte{0xc0, 0xa8, 0x00, 0x01},
			},
			{
				AttrFlags:    0x40,
				AttrTypeCode: update.LOCAL_PREF,
				AttrValue:    []byte{0x00, 0x00, 0x00, 0x64},
			},
		},
		[]netip.Prefix{
			netip.MustParsePrefix("192.168.0.0/24"),
		},
	)
	fmt.Println("before send")

	if err := send_message(s.Conn, sample); err != nil {
		log.Fatalf("failed to write: %v", err)
	}
	fmt.Println("after send")

	b := new(bytes.Buffer)
	if _, err := io.CopyN(b, s.Conn, 19); err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	h := header.Header{}
	binary.Read(b, binary.BigEndian, &h)

	b = new(bytes.Buffer)

	log.Println("length: ", h.Length)
	if _, err := io.CopyN(b, s.Conn, int64(h.Length)-19); err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	msg, err := update.UnMarshal(b.Bytes())
	if err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
	}
	log.Printf("msg: %v", msg)
}

func handle_bgp(ifi net.Interface, active bool) {
	session := Session{
		State:               Idle,
		ConnectRetryCounter: 0,
		ConnectRetryTime:    120 * time.Second,
		HoldTime:            180 * time.Second,
		KeepaliveTime:       60 * time.Second,
		Events:              make(chan Event),
	}

	for {
		log.Println("session state: ", session.State)
		switch session.State {
		case Idle:
			session.Idle(ifi, active)
		case Connect:
			session.Connect()
		case Active:
			session.Active()
		case OpenSent:
			session.OpenSent()
		case OpenConfirm:
			session.OpenConfirm()
		case Established:
			session.Established()
		default:
			log.Fatalf("unknown state: %v", session.State)
		}
		time.Sleep(1 * time.Second)
	}
}

func peers_ifi(active bool) {
	ifis, err := net.Interfaces()
	if err != nil {
		log.Fatalf("failed to get interfaces: %v", err)
	}
	for _, ifi := range ifis {
		if is_loopback(ifi) {
			continue
		}
		log.Printf("sending bgp from %v", ifi.Name)
		go handle_bgp(ifi, active)
	}
	select {}
}
