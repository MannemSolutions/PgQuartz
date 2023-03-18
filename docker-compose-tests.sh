#!/bin/bash
set -e

docker-compose down --remove-orphans && docker rmi pgquartz-builder pgquartz-stolon || echo new or partial install
docker-compose up -d --scale stolon=3
docker exec pgquartz-builder-1 /bin/bash -ic 'cd /host && make build_dlv build_pgquartz'

for ((i=1;i<=3;i++)); do
  echo "pgquartz-stolon-${i}"
  docker exec "pgquartz-stolon-${i}" bash -c '/host/bin/pgquartz.$(uname -m) -c /host/jobs/jobspec1/job.yml' &
  sleep 1
done

echo "All is as expected"
