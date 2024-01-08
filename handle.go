package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/81ueman/local-clos/message"
	"github.com/81ueman/local-clos/message/keepalive"
	"github.com/81ueman/local-clos/message/open"
)

func (s *Session) receiveMessage() {
	for {
		msg, err := message.UnMarshal(s.Conn)
		if err != nil {
			log.Printf("failed to UnMarshal: %v", err)
			s.Cancel()
			return
		}
		log.Printf("msg: %v", msg)
		s.MsgCh <- msg
	}
}

func (s *Session) Idle() {
	var conn net.Conn
	var err error
	if s.ActiveMode {
		conn, err = start_tcp(s.Ifi)
	} else {
		conn, err = wait_tcp(s.Ifi)
	}
	if err != nil {
		log.Printf("failed to handle tcp connection: %v", err)
		s.Cancel()
		s.Events <- TcpConnectionFails
		return
	}
	s.Conn = conn
	go s.receiveMessage()
	s.Events <- Tcp_CR_Acked

	log.Printf("Idle: %v", s.State)
	if s.ActiveMode {
		s.State = Connect
	} else {
		s.State = Active
	}
}

func (s *Session) Connect() {
	event := <-s.Events
	switch event {
	case Tcp_CR_Acked:
		open_msg := open.New(4, s.AS, 180, 0)
		err := message.Send_message(s.Conn, open_msg)
		if err != nil {
			s.Cancel()
			log.Printf("failed to write: %v", err)
			return
		}
		s.State = OpenSent
	case TcpConnectionFails:
		log.Printf("failed to connect to peer")
		s.State = Idle
	default:
		log.Printf("unknown event: %v", event)
		s.Cancel()
	}
}

func (s *Session) Active() {
	event := <-s.Events
	switch event {
	case Tcp_CR_Acked:
		open_msg := open.New(4, 65000, 180, 0)
		err := message.Send_message(s.Conn, open_msg)
		if err != nil {
			log.Fatalf("failed to send message: %v", err)
		}
		s.State = OpenSent
	case TcpConnectionFails:
		s.State = Active
		s.Cancel()
	default:
		log.Printf("unknown event: %v", event)
		s.Cancel()
	}
}

func (s *Session) OpenSent() {
	log.Println("s conn: ", s.Conn.RemoteAddr().String())
	msg := <-s.MsgCh
	msgtype, err := message.Type(msg)
	if err != nil {
		log.Printf("failed to get type: %v", err)
		return
	}
	if msgtype != message.MsgTypeOpen {
		log.Printf("expected an open message, but got: %v", msgtype)
	}
	log.Printf("msg: %v", msg)
	s.State = OpenConfirm
}

func (s *Session) OpenConfirm() {
	keepalive := keepalive.New()
	err := message.Send_message(s.Conn, keepalive)
	if err != nil {
		log.Fatalf("failed to write: %v", err)
		s.Cancel()
		return
	}
	msg := <-s.MsgCh
	msgtype, err := message.Type(msg)
	if err != nil {
		log.Printf("failed to get type: %v", err)
		s.Cancel()
		return
	}
	if msgtype != message.MsgTypeKeepalive {
		log.Printf("expected a keepalive message, but got: %v", msgtype)
		s.Cancel()
		return
	}
	log.Printf("msg: %v", msg)
	s.AdjRibCh <- s.AdjRIBsIn
	s.State = Established
}

func (s *Session) Established() {
	select {
	case locrib := <-s.LocRibCh:
		log.Printf("ribAdj: %v", locrib)
	case msg := <-s.MsgCh:
		msgtype, err := message.Type(msg)
		if err != nil {
			log.Printf("failed to get type: %v", err)
			s.Cancel()
			return
		}
		if msgtype != message.MsgTypeUpdate {
			log.Printf("expected an update message, but got: %v", msgtype)
			s.Cancel()
			return
		}
		log.Printf("msg: %v", msg)
	}
}

func handle_bgp(ctx context.Context, cancel context.CancelFunc, ifi net.Interface, active bool, AS uint16, RibAdjInCh chan RibAdj, LocRibCh chan RibAdj) {
	netipIp, err := localNetipIp(ifi)
	if err != nil {
		log.Fatalf("failed to get local netip ip: %v", err)
		cancel()
		return
	}
	s := Session{
		State:               Idle,
		ConnectRetryCounter: 0,
		ConnectRetryTime:    120 * time.Second,
		HoldTime:            180 * time.Second,
		KeepaliveTime:       60 * time.Second,
		Events:              make(chan Event, 2),
		ActiveMode:          active,
		Ifi:                 ifi,
		NetipAddr:           netipIp,
		AS:                  AS,
		MsgCh:               make(chan message.Message, 10), //magic number to be determined
		AdjRibCh:            RibAdjInCh,
		LocRibCh:            LocRibCh,
		Ctx:                 ctx,
		Cancel:              cancel,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("handle_bgp finished")
			s.Conn.Close()
			return
		default:
			log.Println("session state: ", s.State)
			switch s.State {
			case Idle:
				s.Idle()
			case Connect:
				s.Connect()
			case Active:
				s.Active()
			case OpenSent:
				s.OpenSent()
			case OpenConfirm:
				s.OpenConfirm()
			case Established:
				s.Established()
			default:
				log.Fatalf("unknown state: %v", s.State)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func peers_ifi(active bool, AS uint16) []Peer {
	ifis, err := net.Interfaces()
	peers := make([]Peer, 0, len(ifis))
	if err != nil {
		log.Fatalf("failed to get interfaces: %v", err)
	}
	for _, ifi := range ifis {
		if is_loopback(ifi) {
			continue
		}
		log.Printf("sending bgp from %v", ifi.Name)
		ctx, cancel := context.WithCancel(context.Background())
		RibAdjInCh := make(chan RibAdj)
		LocRibCh := make(chan RibAdj)

		peer := Peer{
			RibAdjIn:   make(RibAdj),
			RibAdjInCh: RibAdjInCh,
			LocRibCh:   LocRibCh,
		}
		peers = append(peers, peer)
		go handle_bgp(ctx, cancel, ifi, active, AS, RibAdjInCh, LocRibCh)
	}
	return peers
}
