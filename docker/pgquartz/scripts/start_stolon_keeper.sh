#!/bin/bash
SCRIPTDIR=$(dirname $0)
eval $($SCRIPTDIR/config_stolon.sh | sed -e 's/#.*//' -e '/[a-zA-Z0-9]/!d' -e 's/^/export /')
if [ "$(id -un)" != postgres ]; then
  mkdir -p "${STKEEPER_DATA_DIR}" "${STKEEPER_WAL_DIR}"
  chown postgres: "${STKEEPER_DATA_DIR}" "${STKEEPER_WAL_DIR}"
  su - postgres $0
  exit
fi
stolon-keeper
