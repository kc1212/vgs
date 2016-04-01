package main

import (
	"fmt"
)

import "github.com/kc1212/vgs/model"

func main() {
	model.RunResMan(2, 2, "localhost:3100", "localhost:3333")
	fmt.Println("test")
}
