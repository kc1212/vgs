package main

import (
	"fmt"
)

import "github.com/kc1212/vgs/model"

func main() {
	model.RunResMan(2, ":1234")
	fmt.Println("test")
}
