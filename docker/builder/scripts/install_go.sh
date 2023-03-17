#!/bin/bash
set -e
dnf update -y
dnf groupinstall -y "Development Tools"
dnf install -y bind-utils make git iproute

PKGARCH=$(uname -m | sed 's/aarch64/arm64/;s/x86_64/amd64/')
cd $(mktemp -d) && \
curl -L https://go.dev/dl/go${GOVER}.linux-${PKGARCH}.tar.gz -o go${GOVER}.linux-${PKGARCH}.tar.gz && \
rm -rf /usr/local/go && tar -C /usr/local -xzf go${GOVER}.linux-${PKGARCH}.tar.gz && \
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/bashrc
