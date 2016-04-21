package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"time"
)

import "github.com/kc1212/virtual-grid/model"

func main() {
	addr := flag.String("addr", "localhost:3000", "address:port of the grid scheduler")
	jobsCount := flag.Int("count", 1, "the number of jobs to add")
	duration := flag.Int64("duration", 0, "the duration for the jobs (default is a random value)")
	nodeType := flag.String("type", "gs", "add job on \"gs\" or \"rm\"")
	flag.Parse()

	if *nodeType != "gs" && *nodeType != "rm" {
		flag.PrintDefaults()
		return
	}

	rand.Seed(time.Now().UTC().UnixNano())
	jobs := make([]model.Job, *jobsCount)
	for i := range jobs {
		jobs[i].ID = rand.Int63()
		var str string
		if *duration == 0 {
			str = fmt.Sprintf("%vs", rand.Intn(10)+1)
		} else {
			str = fmt.Sprintf("%vs", *duration)
		}

		duration, e := time.ParseDuration(str)
		if e != nil {
			log.Panic(e)
		}
		jobs[i].Duration = duration
		jobs[i].StartTime = time.Now()
	}

	reply := -1
	remote, e := rpc.DialHTTP("tcp", *addr)
	if e != nil {
		log.Printf("Node %v is not online, make sure to use the correct address? %v\n", *addr, e.Error())
		return
	}

	if *nodeType == "gs" {
		if e := remote.Call("GridSdr.AddJobsViaUser", &jobs, &reply); e != nil {
			log.Printf("Remote call GridSdr.AddJobsViaUser failed on %v, %v\n", *addr, e.Error())
			return
		}
	} else if *nodeType == "rm" {
		if e := remote.Call("ResMan.AddJobsViaUser", &jobs, &reply); e != nil {
			log.Printf("Remote call ResMan.AddJobsViaUser failed on %v, %v\n", *addr, e.Error())
			return
		}
	}
}
