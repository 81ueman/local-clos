package main

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/81ueman/local-clos/message"
)

type State string

const (
	Idle        State = "Idle"
	Connect     State = "Connect"
	Active      State = "Active"
	OpenSent    State = "OpenSent"
	OpenConfirm State = "OpenConfirm"
	Established State = "Established"
)

type Session struct {
	State               State
	ConnectRetryCounter int
	ConnectRetryTimer   time.Timer
	ConnectRetryTime    time.Duration
	HoldTimer           time.Timer
	HoldTime            time.Duration
	KeepaliveTimer      time.Timer
	KeepaliveTime       time.Duration
	ActiveMode          bool
	Ifi                 net.Interface
	NetipAddr           netip.Addr
	Conn                net.Conn
	MsgCh               chan message.Message
	Addr                net.Addr
	Events              chan Event
	AS                  uint16
	AdjRIBsIn           RibAdj
	AdjRIBsOut          RibAdj
	AdjRibCh            chan<- RibAdj
	LocRibCh            <-chan RibAdj
	Ctx                 context.Context
	Cancel              context.CancelFunc
}

type Event string

const (
	ManualStart                 Event = "ManualStart"
	ManualStop                  Event = "ManualStop"
	ConnectRetryCounter_Expires Event = "ConnectRetryCounter_Expires"
	HoldTimer_Expires           Event = "HoldTimer_Expires"
	KeepaliveTimer_Expires      Event = "KeepaliveTimer_Expires"
	DelayOpenTimer_Expires      Event = "DelayOpenTimer_Expires"
	Tcp_CR_Acked                Event = "Tcp_CR_Acked"
	TcpConnectionConfirmed      Event = "TcpConnectionConfirmed"
	TcpConnectionFails          Event = "TcpConnectionFails"
	BGPOpen                     Event = "BGPOpen"
	BGPHeaderErr                Event = "BGPHeaderErr"
	BGPOpenMsgErr               Event = "BGPOpenMsgErr"
	NotifMsgVerErr              Event = "NotifMsgVerErr"
	NotifMsg                    Event = "NotifMsg"
	KeepAliveMsg                Event = "KeepAliveMsg"
	UpdateMsg                   Event = "UpdateMsg"
	UpdateMsgErr                Event = "UpdateMsgErr"
)
