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

func (s *Session) Idle(ifi net.Interface, active bool) {
	var conn net.Conn
	var err error
	if active {
		conn, err = start_tcp(ifi)
	} else {
		conn, err = wait_tcp(ifi)
	}
	if err != nil {
		log.Printf("failed to handle tcp connection: %v", err)
		s.Cancel()
		s.Events <- TcpConnectionFails
		return
	}
	s.Conn = conn
	s.Events <- Tcp_CR_Acked

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
	log.Println("session conn: ", s.Conn.RemoteAddr().String())
	msg, err := message.UnMarshal(s.Conn)
	if err != nil {
		log.Printf("failed to UnMarshal: %v", err)
		return
	}
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
	msg, err := message.UnMarshal(s.Conn)
	if err != nil {
		log.Printf("failed to UnMarshal: %v", err)
		s.Cancel()
		return
	}
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
	s.State = Established
}

func (s *Session) Established() {
	select {}
}

func handle_bgp(ctx context.Context, cancel context.CancelFunc, ifi net.Interface, active bool) {
	session := Session{
		State:               Idle,
		ConnectRetryCounter: 0,
		ConnectRetryTime:    120 * time.Second,
		HoldTime:            180 * time.Second,
		KeepaliveTime:       60 * time.Second,
		Events:              make(chan Event),
		Ctx:                 ctx,
		Cancel:              cancel,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("handle_bgp finished")
			return
		default:
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
}

func peers_ifi(active bool) ([]context.Context, []context.CancelFunc) {
	ifis, err := net.Interfaces()
	ctxs := make([]context.Context, len(ifis))
	cancels := make([]context.CancelFunc, len(ifis))
	if err != nil {
		log.Fatalf("failed to get interfaces: %v", err)
	}
	for _, ifi := range ifis {
		if is_loopback(ifi) {
			continue
		}
		log.Printf("sending bgp from %v", ifi.Name)
		ctx, cancel := context.WithCancel(context.Background())
		ctxs = append(ctxs, ctx)
		cancels = append(cancels, cancel)
		go handle_bgp(ctx, cancel, ifi, active)
	}
	return ctxs, cancels
}
