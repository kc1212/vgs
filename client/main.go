package main

import (
	"../model"
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"time"
)

// TODO add interactive option
func main() {
	cli()
}

func cli() {
	addr := flag.String("addr", "localhost:3000", "address:port of the grid scheduler")
	jobsCount := flag.Int("count", 1, "the number of jobs to add")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())
	jobs := make([]model.Job, *jobsCount)
	for i := range jobs {
		jobs[i].ID = rand.Int63()
	}

	reply := -1
	remote, e := rpc.DialHTTP("tcp", *addr)
	if e != nil {
		log.Printf("Node %v not online (DialHTTP)\n", addr)
	}
	if e := remote.Call("GridSdr.AddJobs", &jobs, &reply); e != nil {
		log.Printf("Remote call GridSdr.RecvJobs failed on %v, %v\n", addr, e.Error())
	}
}

func interactive() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
