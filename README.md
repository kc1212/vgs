# Virtual Grid System [![Build Status](https://travis-ci.org/kc1212/virtual-grid.svg?branch=master)](https://travis-ci.org/kc1212/virtual-grid) [![Go Report Card](https://goreportcard.com/badge/github.com/kc1212/virtual-grid)](https://goreportcard.com/report/github.com/kc1212/virtual-grid)

## Assumptions
* We assume the crash failure model, i.e. no Byzantine failure or omission failure.
* If a message is delivered, it should contain the intended data, i.e. there does not exist an active attacker that modifies messages in the network.
* The message delays cannot be arbitrarily long.

## Architecture
### User
* The user submits jobs to either GS or RM.
* The request should fail if the job is not accepted so that the user can try again.

### Grid Scheduler (GS)
* GSs can be located on separate networks.
* There exist a leader GS that runs a greedy scheduling algorithm and sends jobs to RMs.
* The leader node is elected using the Bully algorithm.
* Non-leader node should poll the leader to check whether it is online, if the leader crashes then the remaining nodes must run the Bully algorithm again to elect a new leader.
* We aim to achieve strong consistency.
* Jobs have total ordering.
* Jobs are replicated between all the GSs, this is achieved via a modified version of Ricart-Agrawala algorithm.
* If a GS wishes to modify the job queue (add, update or delete), it would request for the critical section. Once it's in the critical section it would broadcast the modification request (using RPC) so that the job queue is consistent across all GSs.
* The original Ricart-Agrawala algorithm would wait indefinitely if a node crashes and does not respond, we modify the algorithm by introducing a timeout to identify crashes so the algorithm can continue to run. It would be tricky to tune the timeout because the GSs may be on different geographical locations hence different message delays.
* The GSs should also maintain a list of jobs that are running (i.e. submitted to a RM), and check whether any of the RMs responsible for those jobs are online.
* If the RM failed, the GS should re-schedule those jobs that were originally on the failed RM to a different RM (see [Job Queue][]).
* The GS network should function even when all of them fail except one.

### Job Queue
* We keep two job queues, one is for incoming jobs `incomingJobs`, i.e. submitted by the user, the other is for scheduled jobs `scheduledJobs`, i.e. scheduled by the GS or RM.
* When there are items in `incomingJobs`, the GS will try to schedule them if there is a RM with enough capacity.
* At the same time the GS will record the responsible RM for every job.
* Scheduled jobs are deleted from `incomingJobs` and added to `scheduledJobs`.
* A job is removed from `scheduledJobs` when the RM announces that the job is completed.
* The GS will poll all the responsible RMs, if they go offline, the GS will re-schedule the job.

### Resource Manager (RM)
* When a job is received from the user, the RM would check whether any of its nodes are free. If a free node exists then the job is assigned to that node, otherwise the job is send back to a random GS that is online for load balancing.
* When a job is received from a GS, the RM must put it into its job queue and process it.
* Once the job is completed, the RM notifies a random GS that is online about its completion, and the GS should delete that job.

### Discovery Server
* There exist a discovery/bootstrap server that is needed to build the network.
* It maintains a list of nodes (GS or RM) that are or was online.
* The discovery server has a static IP address and we assume it cannot crush during the initial bootstrapping process.
* The discovery server is not needed once the bootstrapping process is completed.
* When a node, say `X`, comes online, it sends a message to the discovery server of its presence.
* The discovery server should reply `X` with a list of all the other nodes currently in the system.
* Upon receiving the list of nodes, `X` sends a message to every other node in the list so that those nodes knows about `X`'s existance.
* Nodes sends a "I'm alive" message to the discovery server (if it's online) every 10 seconds.
* If the discovery server fails to receive a "I'm alive" message from some node in 20 seconds, it removes that node from the list.

## Diagram
![Diagram](/diagram.png?raw=true "Diagram")

## Implementation Notes
* RPC is used for all forms of communication.
* Ricart-Agrawala implementation according to pseudocode of [these](http://www2.imm.dtu.dk/courses/02222/Spring_2011/W9L2/Chapter_12a.pdf) slides.
* The Bully Algorithm is implemented by following "Distributed Systems - Principals and Paradigms" by Tannenbaum.

## Building and Running
* Type `make` to build everything.
* Type `make x` to build component `x`, where `x` can be `cli`, `discosrv`, `gridsdr` or `resman`.
* Please start the discovery server - `discosrv` first, before starting `gridsdr` or `resman`. This is not a hard requirement, it's just easier than managing config files.
* User interaction is done through the `cli` executable.
