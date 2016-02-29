package model

import "math/rand"

// the grid scheduler
type GS struct {
	addr     string
	id       int
	others   Ring // other GS's
	clusters []string
	leader   string
}

type Ring []string

func InitGS(addr string) GS {
	id := rand.Int()
	others := *new([]string)   // TODO read from config file or have bootstrapper
	clusters := *new([]string) // TODO see above
	leader := ""
	return GS{addr, id, others, clusters, leader}
}

// TODO how should the user submit request
// via REST API or RPC call from a client?
