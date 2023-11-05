package main

import (
	"fmt"
	"net"
)

func main() {
	ifis, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, ifi := range ifis {
		fmt.Println(ifi.Name, ifi.HardwareAddr, ifi.Flags, ifi.MTU)
		addrs, err := ifi.Addrs()
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			fmt.Println(addr.Network(), addr.String())
		}
		fmt.Println("")
	}
}
