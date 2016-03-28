package main

import (
	"./model"
	"fmt"
)

func main() {
	model.StartResMan(2, ":1234")
	fmt.Println("test")
}
