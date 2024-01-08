// Description: local-closのメインプログラム
package main

import (
	"flag"
	"fmt"
	"log"
)

const ACTIVE bool = true
const PASSIVE bool = false

func main() {
	mode := flag.String("mode", "active", "active or passive")
	//TODO: modeのvalidation欲しいな
	AS := flag.Uint("as", 65000, "AS number")

	flag.Parse()
	// using the ipv6 link-local address will be interesting
	go maintain_arptable()

	var peers []Peer
	if *mode == "active" {
		peers = peers_ifi(ACTIVE, uint16(*AS))
	} else if *mode == "passive" {
		peers = peers_ifi(PASSIVE, uint16(*AS))
	} else {
		log.Fatal("usage: ./local-clos [active|passive]")
	}
	for _, peer := range peers {
		fmt.Printf("%v\n", peer)
	}
	LocRib := LocRib{
		adjBest: make(RibAdj),
		peers:   peers,
	}
	for {
		LocRib.Handle()
	}
}
