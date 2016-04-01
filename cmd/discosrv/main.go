package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

import "github.com/kc1212/vgs/discosrv"

func main() {
	ds := discosrv.DiscoSrv{}

	defaultAddr, e := os.Hostname()
	if e != nil {
		log.Panic("Failed to get hostname")
	}
	defaultAddr = defaultAddr + ":" + "3333"

	discorvAddr := flag.String("addr", defaultAddr, "hostname:port for the DiscoSrv")

	flag.Parse()

	ds.Run(*discorvAddr)
	fmt.Println("test")
}
