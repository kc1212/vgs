# Virtual Grid System

## Assumptions
* We assume the crash failure model, i.e. not the Byzantine failure model.
* The network does not need to be FIFO.
* If a message is delivered, it should contain the intended data, i.e. we assume consistency.

## Architecture
### Grid Scheduler (GS)
* GSs can be located on separate networks.
* There exist a leader GS that runs a (greedy?) scheduling algorithm and sends jobs to RMs.
* The leader is elected using the Bully algorithm.
* Non-leader nodes should constantly ping the leader to check whether it is still online, if the leader crashes then the remaining node must run the Bully algorithm again to elect a new leader.
* Jobs are replicated between all the GSs, we achieve it through a modified version of Ricart-Agrawala algorithm.
* Each node contains a copy of the job queue, the node that is in the critical section sends add or delete messages to all the other nodes to keep the job queue consistent.
* The original Ricart-Agrawala algorithm would wait indefinitely if a crash occurs, we modify the algorithm by introducing a timeout (or pinging?) to identify crashes so the algorithm can continue to run.

### Resource Manager (RM)
* TODO

## Implementation
* TODO

## Building and Running
* TODO
