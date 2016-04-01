package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

import "github.com/kc1212/vgs/model"

func main() {
	// default addr:port
	defaultAddr, e := os.Hostname()
	if e != nil {
		log.Panic("Failed to get hostname")
	}
	defaultAddr = defaultAddr + ":" + "3000"

	n := flag.Int("nodes", 32, "number of workers")
	id := flag.Int("id", 0, "id of the ResMan")
	addr := flag.String("addr", defaultAddr, "hostname:port for this ResMan")
	discosrvAddr := flag.String("discosrv", "localhost:3333", "address of discovery server")

	flag.Parse()

	model.Run(*n, *id, *addr, *discosrvAddr)
	fmt.Println("test")
}
