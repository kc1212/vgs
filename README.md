# Virtual Grid System

## Assumptions
* We assume the crash failure model, i.e. no Byzantine failure or omission failure.
* The network does not need to be FIFO.
* If a message is delivered, it should contain the intended data, i.e. there does not exist an active attacker that modifies messages in the network.
* The message delays cannot be arbitrarily long.

## Architecture
### User
* The user submits jobs to either GS or RM (RPC or REST?).
* The request should fail if the job is not accepted so that the user can try again.

### Grid Scheduler (GS)
* GS's can be located on separate networks.
* There exist a leader GS that runs a greedy scheduling algorithm and sends jobs to RM's.
* The leader node is elected using the Bully algorithm.
* Non-leader node should poll the leader to check whether it is online, if the leader crashes then the remaining nodes must run the Bully algorithm again to elect a new leader.
* Jobs are replicated between all the GS's, this is achieved via a modified version of Ricart-Agrawala algorithm.
* If a GS wishes to modify the job queue (add, update or delete), it would request for the critical section. Once it's in the critical section it would broadcast the modification request (using RPC) so that the job queue is consistent across all GS's.
* The original Ricart-Agrawala algorithm would wait indefinitely if a node crashes and does not respond, we modify the algorithm by introducing a timeout to identify crashes so the algorithm can continue to run. It would be tricky to tune the timeout because the GS's may be on different geographical locations hence different message delays.
* The GS's should also maintain a list of jobs that are running (i.e. submitted to a RM), and check whether any of the RM's responsible for those jobs are online.
* If the RM failed, the GS should re-schedule those jobs that were originally on the failed RM to a different RM (see [Job Queue][]).
* The GS network should function even when all of them fail except one.

### Job Queue
The following snippet shows the job queue.
```
type JobQueue []Job
type Job struct {
    id        int      // or string, must be unique
    duration  int64    // in seconds
    history   []string
    submitted bool
}
```
* The interesting part is `history`, which tracks the sequence of GS and/or RM that the job has passed through.
* If the last entry in the history is a RM and the `submitted` flag is true, then it would imply that the RM is processing the job, so the current leader should check whether that RM is online otherwise re-schedule the job to another RM.
* A job is only removed from the queue when the RM announces that the job is completed.

### Resource Manager (RM)
* When a job is received from the user, the RM would check whether any of its nodes are free. If a free node exists then the job is assigned to that node, otherwise the job is send back to a random GS that is online for load balancing.
* When a job is received from a GS, the RM must put it into its job queue and process it.
* Once the job is completed, the RM notifies a random GS that is online about its completion, and the GS should delete that job.

## Diagram
![Diagram](/DS_Diagram.png?raw=true "Diagram")

## Implementation
* TODO

## Building and Running
* TODO
