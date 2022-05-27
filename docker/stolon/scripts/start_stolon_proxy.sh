#!/bin/bash
if [ "$(id -un)" != postgres ]; then
  su - postgres $0
  exit
fi
SCRIPTDIR=$(dirname $0)
eval $($SCRIPTDIR/config_stolon.sh | sed -e 's/#.*//' -e '/[a-zA-Z0-9]/!d' -e 's/^/export /')

/usr/local/bin/stolonctl status || /usr/local/bin/stolonctl init -y

stolon-proxy
