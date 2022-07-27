# Scheduling

## Systemd/Timer
When scheduling jobs on VM Deployed clusters, the PgQuartz team advices using the Systemd/Timer units over cron entries.
With the Systemd/Timer implementation, every job (defined by a job config file) is defined as a separate service unit, and is scheduled with a systemd timer unit.
This means that every job consists of:
- A [job].yml file which defines the job
   - Including any required scripts that are part of this job, but they can be shared across jobs
- A [job].service unit file (`oneoff`) in /etc/systemd/system/ which runs PgQuartz as the `pgquartz` user specifying the path to the job.yml file
- A [job].timer unit file in /etc/systemd/system/ which triggers the service to run the job at specified moments in time

The beauty of the setup is that every job gets to be a uniquely identifiable unit, and all metadata on that job (last execution result, logging) gets to be a part of that unit.
Moreover, generic stuff, like log space recycling, runs of all jobs, etc. can be managed by the generic solution that systemd has to offer.

### Defining a systemd service and timer
Defining a PgQuartz job can be easily achieved by creating a service and timer using below templates (modify as required for your use case):

Service: `/etc/systemd/system/pgquartz_myjob.service`:
```
# This service unit is for running a pgQuartz job 'as a service'.
# PgQuartz job services can be triggered by systemd timers for scheduling

[Unit]
Description=Run a pgQuartz job

[Service]
Type=oneshot
Environment=PGQUARTZ_CONFIG=/etc/pgquartz/jobs/myjob.yml
ExecStart=/usr/local/bin/pgquartz
User=pgquartz

[Install]
WantedBy=multi-user.target
```

Timer: `/etc/systemd/system/pgquartz_myjob.timer`:
```
# This timer unit is for scheduling a pgQuartz job 'as a service'.

[Unit]
Description=Run the pgQuartz myjob Daily at quarter past 11 pm.
Requires=pgquartz_myjob.service

[Timer]
Unit=pgquartz_myjob.service
OnCalendar=*-*-* 23:15:00

[Install]
WantedBy=timers.target
```

## Cron
The de facto solution for scheduling scripts has always been cron, and of course cron is also supported by PgQuartz.
Using cron has some minimal cons against systemd, but they are negligible in most cases:
- you need to manage log files.
   - Easiest is to specify a log folder as logFile, and issue a logrotate file for cleaning
- you need to set up mail for alerting
   - PgQuartz logs errors to stderr, so redirecting stdout to /dev/null allows for only being mailed on errors
   - setting up a logFile location allows for seeing a combined stderr and stdout error trail when needed

Scheduling a pgQuartz job with cron gets to be as easy as:

Cron: `/etc/cron.d/pgquartz`:
```
MAILTO = me@example.com
15 23 * * * pgquartz /usr/local/bin/pgquartz -c /etc/pgquartz/jobs/myjob.yml >/dev/null
```

LogRotate: `/etc/logrotate.d/pgquartz`:
```
/var/log/pgquartz/pgquartz.log {
    missingok
    ifempty
    daily
    compress
    rotate 7
}
```

## Kubernetes cron
PgQuartz is designed to be run as a Kubernetes cron job, which could be as easy as:
- Create an image with PgQuartz
- Supply the job definition in a ConfigMap
- Schedule a kubernetes cron job to run the Job Pod at required times

We are working on building an image and defining a more verbose runtime example in Issue #40.
