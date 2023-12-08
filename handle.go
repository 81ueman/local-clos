package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/netip"
	"strings"
	"time"

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
	s.Events <- Established_Event
}

func (s *Session) Established() {
	event := <-s.Events
	switch event {
	case Established_Event:
		log.Println("entered established:)")
		prefixes := make([]netip.Prefix, 0)
		ip, err := local_ip(s.Ifi)
		if err != nil {
			log.Fatalf("failed to get local ip: %v", err)
		}
		prefixes = append(prefixes, netip.PrefixFrom(netip.MustParseAddr(ip.String()), 24))
		next_hop := netip.MustParseAddr(strings.Split(s.Conn.LocalAddr().String(), ":")[0])
		if err != nil {
			log.Fatalf("failed to marshal: %v", err)
		}
		update_msg := update.New(
			[]netip.Prefix{},
			[]update.Attrer{
				update.IGP,
				update.ASPathSegment{
					PathSegmentType: update.AS_SEQUENCE,
					AS:              []update.AS{65001},
				},
				update.NEXT_HOP(next_hop),
				update.LOCAL_PREF(100),
			},
			prefixes,
		)
		if err := send_message(s.Conn, update_msg); err != nil {
			log.Fatalf("failed to send message: %v", err)
		}
		fmt.Println("after send")
		msg, err := update.Read(s.Conn)
		if err != nil {
			log.Fatalf("failed to read: %v", err)
		}
		s.LocRibIn = make(LocRib)
		for _, prefix := range msg.NetworkLayerReachable {
			s.LocRibIn[prefix] = msg.PathAttrs
		}

		log.Printf("msg: %v", msg)
	default:
		log.Println("living in established state:)")
	}

}

func handle_bgp(ifi net.Interface, active bool) {
	session := Session{
		State:               Idle,
		ConnectRetryCounter: 0,
		ConnectRetryTime:    120 * time.Second,
		HoldTime:            180 * time.Second,
		KeepaliveTime:       60 * time.Second,
		Ifi:                 ifi,
		Events:              make(chan Event, 10),
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
