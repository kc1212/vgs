package main

import (
	"flag"
)

func main() {
	id := flag.Int("id", 0, "id of the grid scheduler")
	n := flag.Int("n", 0, "number of  grid schedulers")
	flag.Parse()
}
