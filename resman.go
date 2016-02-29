package main

import "fmt"
import "github.com/kc1212/vgs/model"

func main() {
	resMan := model.InitResMan(2)
	resMan.PrintStatus()
	fmt.Println("test")
}
