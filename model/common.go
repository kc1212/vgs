package model

import (
	"log"
	"strconv"
	"strings"
)

type MsgType int

const (
	ElectionMsg MsgType = iota
	CoordinateMsg
	CritSectionMsg
)

func idFromAddr(addr string, basePort int) int {
	tmp := strings.Split(addr, ":")
	port, e := strconv.Atoi(tmp[len(tmp)-1])
	if e != nil {
		log.Panic("idFromAddr failed", e)
	}
	return port - basePort
}
