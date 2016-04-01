package main

import (
	"flag"
	"log"
	"net"
	"os"
)

import "github.com/kc1212/vgs/model"

func main() {
	// default addr:port
	defaultAddr, e := os.Hostname()
	if e != nil {
		log.Panic("Failed to get hostname")
	}
	defaultAddr = net.JoinHostPort(defaultAddr, "3000")

	n := flag.Int("nodes", 32, "number of workers")
	id := flag.Int("id", 0, "id of the ResMan")
	addr := flag.String("addr", defaultAddr, "hostname:port for this ResMan")
	discosrvAddr := flag.String("discosrv", "localhost:3333", "address of discovery server")

	flag.Parse()

	model.RunResMan(*n, *id, *addr, *discosrvAddr)
}
