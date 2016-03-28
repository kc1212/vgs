package main

import (
	"../model"
	"log"
	"os"
	"strconv"
)

func main() {
	id, e := strconv.Atoi(os.Args[1])
	if e != nil {
		log.Fatalf("id argument incorrect")
	}

	n, e := strconv.Atoi(os.Args[2])
	if e != nil {
		log.Fatalf("n argument incorrect")
	}

	gs := model.InitGridSdr(id, n, 3000, "localhost:")
	gs.Run()
}
