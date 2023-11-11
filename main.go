// Description: local-closのメインプログラム
package main

import (
	"errors"
	"log"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/mdlayher/arp"
)

/*
ifiがループバックインターフェースかどうかを判定する
*/
func is_loopback(ifi net.Interface) bool {
	return ifi.Flags&net.FlagLoopback != 0
}

// ifiからIPv4アドレスを取得する
// arpパッケージで使うことを見越して返り値をnetip.Addrにしている
func get_ipv4(ifi net.Interface) (netip.Addr, error) {
	addrs, err := ifi.Addrs()
	if err != nil {
		log.Fatalf("failed to get addrs: %v", err)
	}
	ip := netip.IPv4Unspecified()
	for _, addr := range addrs {
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
	if ip == netip.IPv4Unspecified() {
		return netip.IPv4Unspecified(), errors.New("no ipv4 addr found")
	}
	return ip, nil
}

// ifiからGARPを送信する
// 対向機器に自分自身のIPアドレスを通知するために使用しているだけなので
// GARPである必要はないがまあ便利なので
func send_garp(ifi net.Interface) {
	ip, err := get_ipv4(ifi)
	if err != nil {
		log.Fatalf("failed to get ipv4 addr: %v", err)
	}
	log.Printf("local address is %v", ip)
	log.Printf("hardware address is %v", ifi.HardwareAddr)
	broad, err := net.ParseMAC("ff:ff:ff:ff:ff:ff")
	if err != nil {
		log.Fatalf("failed to parse mac: %v", err)
	}
	packet, err := arp.NewPacket(arp.OperationReply, ifi.HardwareAddr, ip, broad, ip)
	if err != nil {
		log.Fatalf("failed to create arp packet: %v", err)
	}
	log.Printf("packet is %v", packet)
	cli, err := arp.Dial(&ifi)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	err = cli.WriteTo(packet, broad)
	if err != nil {
		log.Fatalf("failed to write: %v", err)
	}
}

// send garp packet from all interfaces to maintain opposite arp table
func maintain_arptable() {
	for {
		ifis, err := net.Interfaces()
		if err != nil {
			log.Fatalf("failed to get interfaces: %v", err)
		}
		for _, ifi := range ifis {
			if is_loopback(ifi) {
				continue
			}
			log.Printf("sending garp from %v", ifi.Name)
			send_garp(ifi)
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: ./local-clos [active|passive]")
	}
	mode := os.Args[1]
	go maintain_arptable()

	if mode == "active" {
		active_mode()
	} else if mode == "passive" {
		passive_mode()
	} else {
		log.Fatal("invalid mode")
	}
}
