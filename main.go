// Description: local-closのメインプログラム
package main

import (
	"log"
	"os"
)

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
