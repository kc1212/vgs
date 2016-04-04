package main

import (
	"flag"
	"net"
)

import "github.com/kc1212/virtual-grid/model"

func main() {
	defaultAddr := net.JoinHostPort("localhost", "3000")

	n := flag.Int("nodes", 32, "number of workers")
	id := flag.Int("id", 0, "id of the ResMan")
	addr := flag.String("addr", defaultAddr, "hostname:port for this ResMan")
	discosrvAddr := flag.String("discosrv", "localhost:3333", "address of discovery server")

	flag.Parse()

	rm := model.InitResMan(*n, *id, *addr, *discosrvAddr)
	rm.Run()
}
