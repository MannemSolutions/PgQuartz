#!/bin/bash
set -e

docker-compose down --remove-orphans #&& docker rmi pgquartz_builder pgquartz_stolon || echo new or partial install
docker-compose up -d --scale stolon=3
docker exec pgquartz_builder_1 /bin/bash -ic 'cd /host && make build_dlv build_pgquartz'

for ((i=1;i<=3;i++)); do
  echo "pgquartz_stolon_${i}"
  docker exec "pgquartz_stolon_${i}" /host/bin/pgquartz.x86_64 -c /host/jobs/jobspec1/job.yml &
  sleep 1
done
exit

docker-compose up -d pgroute66
docker ps -a
assert primary 'host1'
assert primaries '[ host1 ]'
assert standbys '[ host2, host3 ]'

docker exec pgroute66_postgres_2 /entrypoint.sh promote
assert primary ''
assert primaries '[ host1, host2 ]'
assert standbys '[ host3 ]'

docker exec pgroute66_postgres_1 /entrypoint.sh rebuild
assert primary 'host2'
assert primaries '[ host2 ]'
assert standbys '[ host1, host3 ]'

echo "All is as expected"
