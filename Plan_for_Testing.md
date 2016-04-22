## Fault tolerance

### Setting
**# of Grid Schedulers**: 5

**# of Resource Managers**: 20

**# of Worker Nodes**: 32/Resource Manager

**# of Discovery Servers**: 1

**# of Job Requests**: 10,000

---

### Scenario 1 (private network - leader failure)
* The system is being tested in, at least, 2 different machines on a 
local network.

* The _Discovery Server_ is initialized as the first node of the system. 
All _Grid Schedulers_ are initialized right afterwards, alongside the 
_Resource Managers_ (including their worker nodes), using a ```run.sh``` 
script which initializes them with random identities, default number of 
worker nodes (#32) and with appropriate port addresses. The nodes are 
initialized on each of the machines, in such a way,
 so as there are _Grid Schedulers_ and _Resource Managers_ in each of the 
 local machines.

* After making sure that all the nodes are online (through the 
_Discovery Server_'s logs), the job requests are sent through the _cli_
which sends 10.000 job requests to a specific _Grid Scheduler_ (could be
the leader or not), with specific duration (1 second in this scenario).

* As the system processes the job requests (syncing them among the _Grid 
Schedulers_, scheduling them to _Resource Managers_ and keeping track of
the finished jobs), the leader _Grid Scheduler_ is deliberately crashed 
in order to test the system's fault-tolerance.

* In order to bring instability to the system, the leader is deliberately
brought online and offline multiple times. This would test the system's
ability to proceed to elections while processing the job requests, change
leader correctly (the leader should always have the highest id) and continue
scheduling jobs to the _Resource Managers_, keeping track of the jobs which
finished and not create dublicated jobs.

### Scenario 2 (public network - leader failure)
* The system is being tested in 4 different machines on _DigitalOcean_. 
The machines are located in Amsterdam.

* The _Discovery Server_ is initialized as the first node of the system in
a one of the machines. All _Grid Schedulers_ are initialized right 
afterwards, alongside the _Resource Managers_ (including their worker 
nodes), using a ```run.sh``` script which initializes them with random 
identities, default number of worker nodes (#32) and with appropriate 
port addresses. The initializations should take part in the rest of the
machines (not on the machine which runs Discovery Server), having equally 
distributed number of the virtual grid's nodes in each of them.

* After making sure that all the nodes are online (through the 
_Discovery Server_'s logs), the job requests are sent through the _cli_
which sends 10.000 job requests to a specific _Grid Scheduler_ (could be
the leader or not), with specific duration (1 second in this scenario).

* As the system processes the job requests (syncing them among the _Grid 
Schedulers_, scheduling them to _Resource Managers_ and keeping track of
the finished jobs), the leader _Grid Scheduler_ is deliberately crashed 
in order to test the system's fault-tolerance.

* In order to bring instability to the system, the leader is deliberately
brought online and offline multiple times. This would test the system's
ability to proceed to elections while processing the job requests, change
leader correctly (the leader should always have the highest id) and continue
scheduling jobs to the _Resource Managers_, keeping track of the jobs which
finished and not create dublicated jobs.

### Scenario 3 (private network - multiple Grid Schedulers' failure)
* The system is being tested in, at least, 2 different machines on a 
local network.

* The _Discovery Server_ is initialized as the first node of the system. 
All _Grid Schedulers_ are initialized right afterwards, alongside the 
_Resource Managers_ (including their worker nodes), using a ```run.sh``` 
script which initializes them with random identities, default number of 
worker nodes (#32) and with appropriate port addresses. The nodes are 
initialized on each of the machines, in such a way, so as there are 
_Grid Schedulers_ and _Resource Managers_ in each of the local machines.

* After making sure that all the nodes are online (through the 
_Discovery Server_'s logs), the job requests are sent through the _cli_
which sends 10.000 job requests to a specific _Grid Scheduler_ (could be
the leader or not), with specific duration (1 second in this scenario).

* As the system processes the job requests (syncing them among the _Grid 
Schedulers_, scheduling them to _Resource Managers_ and keeping track of
the finished jobs), one of the _Grid Scheduler_ is deliberately crashed 
in order to test the system's fault-tolerance.

* In order to bring instability to the system, each of the the _Grid 
Schedulers_ is crashed deliberately and in abstract order. This would test 
the system's ability to proceed to elections while processing the job 
requests, change leader correctly (the leader should always have the 
highest id) and continue scheduling jobs to the _Resource Managers_, 
keeping track of the jobs which finished and not create dublicated jobs. 
Finally, it will test the ability of the system to work with at least one
_Grid Scheduler_ after the others have failed.

### Scenario 4 (public network - multiple Grid Schedulers' failure)
* The system is being tested in 4 different machines on _DigitalOcean_. 
The machines are located in Amsterdam.

* The _Discovery Server_ is initialized as the first node of the system in
a one of the machines. All _Grid Schedulers_ are initialized right 
afterwards, alongside the _Resource Managers_ (including their worker 
nodes), using a ```run.sh``` script which initializes them with random 
identities, default number of worker nodes (#32) and with appropriate 
port addresses. The initializations should take part in the rest of the
machines (not on the machine which runs Discovery Server), having equally 
distributed number of the virtual grid's nodes in each of them.

* After making sure that all the nodes are online (through the 
_Discovery Server_'s logs), the job requests are sent through the _cli_
which sends 10.000 job requests to a specific _Grid Scheduler_ (could be
the leader or not), with specific duration (1 second in this scenario).

* As the system processes the job requests (syncing them among the _Grid 
Schedulers_, scheduling them to _Resource Managers_ and keeping track of
the finished jobs), one of the _Grid Scheduler_ is deliberately crashed 
in order to test the system's fault-tolerance.

* In order to bring instability to the system, each of the the _Grid 
Schedulers_ is crashed deliberately and in abstract order. This would test 
the system's ability to proceed to elections while processing the job 
requests, change leader correctly (the leader should always have the 
highest id) and continue scheduling jobs to the _Resource Managers_, 
keeping track of the jobs which finished and not create dublicated jobs. 
Finally, it will test the ability of the system to work with at least one
_Grid Scheduler_ after the others have failed.