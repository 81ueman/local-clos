package main

import (
	"log"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/mdlayher/arp"
)

func send_garp(ifi net.Interface) {
	addrs, err := ifi.Addrs()
	if err != nil {
		log.Fatalf("failed to get addrs: %v", err)
	}
	var ip netip.Addr
	for _, addr := range addrs {
		// TODO: parse the IP addr with more robust way
		addr, err := netip.ParseAddr(strings.Split(addr.String(), "/")[0])
		if err != nil {
			log.Printf("failed to parse addr: %v", err)
			continue
		}
		if addr.Is4() {
			ip = addr
			break
		}
	}
	packet, err := arp.NewPacket(arp.OperationReply, ifi.HardwareAddr, ip, ifi.HardwareAddr, ip)
	if err != nil {
		log.Fatalf("failed to create arp packet: %v", err)
	}
	cli, err := arp.Dial(&ifi)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	cli.WriteTo(packet, ifi.HardwareAddr)
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: ./local-clos [active|passive]")
	}
	mode := os.Args[1]
	go func() {
		for {
			ifis, err := net.Interfaces()
			if err != nil {
				log.Fatalf("failed to get interfaces: %v", err)
			}
			for _, ifi := range ifis {
				if ifi.Flags&net.FlagLoopback != 0 {
					continue
				}
				send_garp(ifi)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	if mode == "active" {
		active_mode()
	} else if mode == "passive" {
		passive_mode()
	} else {
		log.Fatal("invalid mode")
	}
}
