package main

import (
	"context"
	"net"
	"time"
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
	Conn                net.Conn
	Events              chan Event
	Ctx                 context.Context
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
