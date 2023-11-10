package main

import (
	"errors"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/81ueman/local-clos/header"
)

func local_addr(ifi net.Interface) (net.IP, error) {
	addrs, err := ifi.Addrs()
	if err != nil {
		log.Fatalf("failed to get addrs: %v", err)
	}
	ip := net.IPv4zero
	for _, addr := range addrs {
		addr := net.ParseIP(strings.Split(addr.String(), "/")[0])
		if addr.To4() != nil {
			ip = addr
			break
		}
	}
	if ip.Equal(net.IPv4zero) {
		return net.IPv4zero, errors.New("no ipv4 addr found")
	}
	return ip, nil
}

func remote_addr(ifi net.Interface) (net.IP, error) {
	out, err := exec.Command("/usr/sbin/ip", "neigh", "show", "dev", ifi.Name).Output()
	// format: ip lladdr MAC STALE|REACHABLE|DELAY|PROBE|FAILED
	if err != nil {
		return net.IPv4zero, errors.New("failed to get arp table")
	}
	rip := strings.Split(string(out), " ")[0]
	return net.ParseIP(rip), nil
}

func active_mode() {
	ifis, err := net.Interfaces()
	if err != nil {
		log.Fatalf("failed to get interfaces: %v", err)
	}
	for _, ifi := range ifis {
		if ifi.Flags&net.FlagLoopback != 0 {
			continue
		}

		local_addr, err := local_addr(ifi)
		if err != nil {
			log.Fatalf("failed to get local addr: %v", err)
		}
		laddr := &net.TCPAddr{
			IP:   local_addr,
			Port: 179,
		}

		rip, err := remote_addr(ifi)
		if err != nil {
			log.Fatalf("failed to get remote addr: %v", err)
		}
		raddr := &net.TCPAddr{
			IP:   rip,
			Port: 179,
		}

		conn, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		header := header.New(0xffff, 0xff)
		for {
			hbytes, err := header.Marshal()
			if err != nil {
				log.Fatal(err)
			}
			conn.Write(hbytes)
			time.Sleep(1 * time.Second)
		}
	}
}
