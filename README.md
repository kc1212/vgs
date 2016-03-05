Virtual Grid System
===================

Assumptions
-----------------
* We assume the crash failure model, i.e. not the Byzantine failure model.

Architecture
------------
* GSs (Grid Schedulers) can be located on separate networks.
* There exist a leader GS that runs a (greedy?) scheduling algorithm and sends jobs to RMs.
* The leader is elected using an election algorithm. TODO which algorithm?
* Jobs are replicated between all the GSs. TODO how is the replication done, causal messages?

Implementation
--------------
* TODO

Building and Running
--------------------
* TODO
