package main

import (
	"flag"
	"net"
)

import "github.com/kc1212/vgs/discosrv"

func main() {
	defaultAddr := net.JoinHostPort("localhost", "3333")
	discorvAddr := flag.String("addr", defaultAddr, "hostname:port for the DiscoSrv")

	flag.Parse()

	ds := discosrv.DiscoSrv{}
	ds.Run(*discorvAddr)
}
