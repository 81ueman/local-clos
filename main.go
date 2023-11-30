// Description: local-closのメインプログラム
package main

import (
	"log"
	"os"
)

const ACTIVE bool = true
const PASSIVE bool = false

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: ./local-clos [active|passive]")
	}
	mode := os.Args[1]
	// using the ipv6 link-local address will be interesting
	go maintain_arptable()

	if mode == "active" {
		peers_ifi(ACTIVE)
	} else if mode == "passive" {
		peers_ifi(PASSIVE)
	} else {
		log.Fatal("usage: ./local-clos [active|passive]")
	}
}
