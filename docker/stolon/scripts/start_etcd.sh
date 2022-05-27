#!/bin/bash
if [ "$(id -un)" != etcd ]; then
  su - etcd $0
  exit
fi
SCRIPTDIR=$(dirname $0)
eval $($SCRIPTDIR/config_etcd.sh | sed -e 's/#.*//' -e '/[a-zA-Z0-9]/!d' -e 's/^/export /')
etcd
