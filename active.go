package main

import (
	"errors"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/81ueman/local-clos/open"
)

var errArptableNotFound = errors.New("no arp table found")

func local_ip(ifi net.Interface) (net.IP, error) {
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
func local_addr(ifi net.Interface) (*net.TCPAddr, error) {
	ip, err := local_ip(ifi)
	if err != nil {
		return &net.TCPAddr{}, err
	}
	laddr := &net.TCPAddr{
		IP:   ip,
		Port: 179,
	}
	return laddr, nil
}

func remote_ip(ifi net.Interface) (net.IP, error) {
	out, err := exec.Command("/usr/sbin/ip", "neigh", "show", "dev", ifi.Name).Output()
	// format: ip lladdr MAC STALE|REACHABLE|DELAY|PROBE|FAILED
	if err != nil {
		return net.IPv4zero, errors.New("failed to get arp table")
	}
	log.Printf("arp table: %v", string(out))
	if len(out) == 0 {
		return net.IPv4zero, errArptableNotFound
	}
	rip := strings.Split(string(out), " ")[0]
	return net.ParseIP(rip), nil
}

func remote_tcpaddr(ifi net.Interface) (*net.TCPAddr, error) {
	rip, err := remote_ip(ifi)
	if err != nil {
		return &net.TCPAddr{}, err
	}
	raddr := &net.TCPAddr{
		IP:   rip,
		Port: 179,
	}
	return raddr, nil
}

func remote_addr_loop(ifi net.Interface) (*net.TCPAddr, error) {
	var raddr *net.TCPAddr
	for {
		var err error
		raddr, err = remote_tcpaddr(ifi)
		if err == errArptableNotFound {
			log.Println("remote addres is not found in the ARP table. Waiting one more second...")
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			log.Printf("failed to get remote addr: %v", err)
			return &net.TCPAddr{}, err
		} else {
			break
		}
	}
	return raddr, nil
}

func send_bgp(ifi net.Interface) {
	laddr, err := local_addr(ifi)
	if err != nil {
		log.Printf("failed to get local addr: %v", err)
		return
	}
	raddr, err := remote_addr_loop(ifi)
	if err != nil {
		log.Printf("failed to get remote addr: %v", err)
		return
	}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for {
		bytes, err := Marshal(open.New(4, 65000, 180, 0))
		if err != nil {
			log.Fatalf("failed to marshal: %v", err)
		}
		_, err = conn.Write(bytes)
		if err != nil {
			log.Fatalf("failed to write: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
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
		go send_bgp(ifi)
	}
	for {
		time.Sleep(1 * time.Second)
	}
}
