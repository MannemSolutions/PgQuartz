#!/bin/bash
set -e

function print_config() {
echo "# [member]
ETCD_NAME=${MYHOSTNAME}
ETCD_DATA_DIR=/var/lib/etcd/stolondebug.etcd
ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
ETCD_ENABLE_V2=True
#
# [cluster]
ETCD_LISTEN_PEER_URLS=http://${MYIP}:2380
ETCD_ADVERTISE_CLIENT_URLS=http://${MYIP}:2379
ETCD_INITIAL_ADVERTISE_PEER_URLS=http://${MYIP}:2380
ETCD_INITIAL_CLUSTER=${ETCD_INITIAL_CLUSTER}
ETCD_INITIAL_CLUSTER_STATE=new
ETCD_INITIAL_CLUSTER_TOKEN=d8bf8cc6-5158-11e6-8f13-3b32f4935bde

# [custom_env_vars]
ETCD_AUTO_COMPACTION_RETENTION=1"
}

CLUSTER_SIZE=${CLUSTER_SIZE:-3}
MYIP=$(ip a | grep -oE 'inet ([0-9]{1,3}\.){3}[0-9]{1,3}' | sed -e '/127\.0\.0\.1/d' -e 's/inet //')
MYHOSTNAME=$(host "${MYIP}" | sed -e 's/.* //' -e 's/\..*//')
MYHOSTNAMEPROFILE=$(echo $MYHOSTNAME | sed 's/[0-9]*$//')
ETCD_INITIAL_CLUSTER=$(for ((i=1;i<=${CLUSTER_SIZE};i++)); do host "${MYHOSTNAMEPROFILE}${i}"; done | awk '{if (NR>1){printf(",")};printf("%s=http://%s:2380",$1,$4)}')

print_config
