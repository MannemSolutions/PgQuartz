# Etcd integration
PgQuartz can integrate with etcd in which case it will lock a key while it is running.
This allows for scheduling jobs across clustered nodes one at a time, where
- any node will lock the key and run the steps
- all other nodes will wait until the node has finished its operations and released the key
- when the key is released the next node will lock the key, run the steps and release the key
- and on and on it goes until either 
  - all nodes have successfully finished running the steps, or
  - the [Jobs timeout](./JOBS.md#timeout) has expired after which all nodes that have not run yet will exit with an error code

## Configuration options

### endpoints
a list of endpoints can be set to point to all instances of etcd.
Every node should be formatted as `hostname:port` (e.a. `server1:2379`).
When not set, this defaults to only one node `localhost:2379`.

### LockKey
The key that will be locked for this job can be configured.
When not set, it defaults to the job name.
**_note_** that the job name is derived from the yaml that defines the job (e.a. `/etc/pgquartz/jobs/job1.yaml` would result in a job name `job1`)

### Lock timeout
While running, PgQuartz automatically refreshes the lock, and when finished, PgQuartz automatically releases the lock.
But there are circumstances where etcd would retain the lock until a preset lock timeout, after which it would release the lock.

Examples of these circumstances are:
- PgQuartz dies a horrible death (e.a. `killall -KILL pgquartz` would be horrible)
- a network partition would occur between PgQuartz and etcd

And when the lock is released:
- the current PgQuartz job fails
- PgQuartz running on another node would be released to run the job

## Example
To explain the 'etcd integration' consider the following config:

### Example config
```
steps:
  step 1:
    commands:
      - name: Run command 1.1
        type: shell
        inline: "sleep 10"
      - name: Run command 1.2
        type: shell
        inline: "date +%s > /tmp/beenhere.txt"
etcdConfig:
  endpoints:
    - localhost:2379
  lockKey: awesomeJob1
  lockTimeout: 1m
timeout: 75s
```

### What does it do under normal circumstances?
As an example the example config would be run as a job on a cluster of 3 nodes in an etcd cluster.
- each node is running etcd on port 2379, and together they form a cluster
- each node has this exact job definition, and PgQuartz is scheduled to run this job at the same time

The following would happen:
1. any node would be first to run PgQuartz and lock the key `awesomeJob1`
   - that node would wait 10 seconds and then add the current epoch (seconds elapsed since January 1, 1970) into a file /tmp/beenhere.txt
   - other nodes wait until the job on the locking node is finished and has released the node
2. When the job on the first node is finished 
   - on this initial node, PgQuartz releases the lock, and wraps up (run checks and exit)
   - one of the other 2 nodes would notice the release and lock the key `awesomeJob1` after which it would
     - wait 10 seconds
     - add the current epoch (seconds elapsed since January 1, 1970) into a file /tmp/beenhere.txt
   - the last node waits until the job on the second node is released
3. When the job on the second node is finished
	- on this second node, PgQuartz releases the lock, and wraps up (run checks and exit)
	- the last node would notice the release and lock the key `awesomeJob1` after which it would
		- wait 10 seconds
		- add the current epoch (seconds elapsed since January 1, 1970) into a file /tmp/beenhere.txt

> **_note_** that:
> - The job has been scheduled to run at the same time
> - The job has actually run after each other one node at a time
> - There is no predetermined order. Probably, ntp drift would decide run order
> - the last line in the /tmp/beenhere.txt files would all differ by 10 seconds (or 11 in some rare cases)

### What happens on issues
If a network partition would occur between the node currently running the job, and the other 2 nodes, the following could happen:
1. If the network partition resolves within 60 seconds, all works out fine
2. If the network partition takes longer:
   - etcd releases the lock
   - the job that was running exits with an error
   - the next node in line would run the job
3. Maybe the second job, and the 3rd job for sure will not succeed due to the job timeout (81 seconds):
   - first node runs (0 - 10) seconds
   - partition hangs for 60 seconds
   - second node runs for 10 seconds (starts 60-70 seconds after job start)
   - if it is started after waiting for 65 seconds or more, it would not finish, but be timed out after 75 seconds
   - If not, it will finish, only to leave 0-5 seconds for node 3 before timing out as well

If any issue occurs (hang, process kill, etc.) either PgQuartz would cleanly release the lock, or the above described behavior would be perceived.