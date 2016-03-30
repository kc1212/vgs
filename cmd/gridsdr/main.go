package main

import (
	"flag"
	"log"
	"os"
)

import "github.com/kc1212/vgs/model"

func main() {
	// default addr:port
	defaultName, e := os.Hostname()
	if e != nil {
		log.Panic("Failed to get hostname")
	}
	defaultName = defaultName + ":" + "3000"

	id := flag.Int("id", 0, "id of the node")
	name := flag.String("addr", defaultName, "hostname:port for this node")
	discosrvAddr := flag.String("discosrv", "localhost:3333", "address of discovery server")

	flag.Parse()

	gs := model.InitGridSdr(*id, *name, *discosrvAddr)
	gs.Run()
}
