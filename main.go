package main

import (
	"log"
	"os"
)

func main() {
	mode := os.Args[1]
	if mode == "active" {
		active_mode()
	} else if mode == "passive" {
		passive_mode()
	} else {
		log.Fatal("invalid mode")
	}
}
