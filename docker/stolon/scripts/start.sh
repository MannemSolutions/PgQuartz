#!/bin/bash
SCRIPTDIR=$(dirname $0)
for SVC in etcd stolon_sentinel stolon_keeper stolon_proxy stolon_init; do
  "${SCRIPTDIR}/start_${SVC}.sh" &
  sleep 1
done
wait
