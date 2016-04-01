package main

import "github.com/kc1212/vgs/discosrv"

func main() {
	ds := discosrv.DiscoSrv{}
	ds.Run("localhost:3333")
}
