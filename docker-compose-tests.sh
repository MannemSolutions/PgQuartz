#!/bin/bash
set -e

docker-compose down && docker rmi pgquartz_pgquartz || echo new install
docker-compose up -d --scale pgquartz=3

#for ((i=1;i<=3;i++)); do
#  echo "stolondebug_stolon_${i}"
#  docker exec -ti "stolondebug_stolon_${i}" /host/stolondebug/stolon/scripts/start.sh
#  sleep 1
#done
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
