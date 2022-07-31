# Introduction
The goal of PgQuartz is to [schedule](SCHEDULING.md) jobs against a Primary/Standby cluster with minimal toil.
We can best describe 'minimal toil' with the alternative: cron and bash scripts.

Although we consider [systemd/timer](SCHEDULING.md#defining-a-systemd-service-and-timer) and [Kubernetes cron](SCHEDULING.md#kubernetes-cron) more mature [scheduling](SCHEDULING.md) solutions, 
PgQuartz is not trying to replace [cron](SCHEDULING.md#cron) as a scheduling solution.
Which means you can just as easily schedule PgQuartz using [cron](SCHEDULING.md#cron) as you could schedule your original bash scripts.

PgQuartz is trying to replace the bash scripts to some extent, mainly for al the generic things like:
- Orchestration of jobs across a Postgres cluster
  - Run on a master, a standby, or on any node
  - Run on only one node, run on every node one at the time, or run on all in parallel
- Reuse of database connections
- Separation of code and configuration
- Separation of run and check

PgQuartz is not trying to replace the actual (application specific) definitions of your (bash/sql) scripts. PgQuartz just brings a framework where they can be defined with minimal toil without being required to define all stuff the framework does for you.

By leveraging PgQuartz, basically your jobs get a clearer definition, less error-prone on generic tasks and more efficient on runtime.

For more information, please see our documentation at [readthedocs.io](https://pgquartz.readthedocs.io/en/latest/).
