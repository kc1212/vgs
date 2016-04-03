package main

import (
	"flag"
	"net"
)

import "github.com/kc1212/vgs/model"

func main() {
	defaultAddr := net.JoinHostPort("localhost", "3000")

	name := flag.String("addr", defaultAddr, "hostname:port for this node")
	id := flag.Int("id", 0, "id of the node")
	discosrvAddr := flag.String("discosrv", "localhost:3333", "address of discovery server")

	flag.Parse()

	gs := model.InitGridSdr(*id, *name, *discosrvAddr)
	gs.Run()
}
