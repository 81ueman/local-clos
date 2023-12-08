package main

import (
	"net"
	"net/netip"
	"strings"
)

func AddrToNetip(addr net.Addr) netip.Addr {
	addr_netip, err := netip.ParseAddr(strings.Split(addr.String(), ":")[0])
	if err != nil {
		panic(err)
	}
	return addr_netip
}

func AddrToPrefix(addr net.Addr, mask int) netip.Prefix {
	addr_netip := AddrToNetip(addr)
	prefix := netip.PrefixFrom(addr_netip, mask)
	return prefix.Masked()
}
