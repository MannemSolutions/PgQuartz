FROM golang:latest AS stolonbuilder
RUN  cd /tmp && \
     git clone https://github.com/sebasmannem/stolon.git && \
     cd stolon && \
     git checkout current && \
     make

FROM rockylinux:8

ENV ETCD_VER=v3.5.2
ENV ETCD_DOWNLOAD_URL=https://github.com/etcd-io/etcd/releases/download

RUN yum update -y && \
    yum install -y bind-utils make git iproute

# Install etcd
RUN cd $(mktemp -d) && \
    PKGARCH=$(uname -m | sed 's/aarch64/arm64/;s/x86_64/amd64/') && \
    curl -L ${ETCD_DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-${PKGARCH}.tar.gz -o etcd-${ETCD_VER}-linux-${PKGARCH}.tar.gz && \
    tar xzvf etcd-${ETCD_VER}-linux-${PKGARCH}.tar.gz && \
    cd etcd-${ETCD_VER}-linux-${PKGARCH} && \
    mv etcd etcdctl etcdutl /usr/local/bin && \
    useradd -mrb /var/lib etcd

# Install stolon
COPY --from=stolonbuilder /tmp/stolon/bin  /usr/local/bin

RUN dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-$(uname -m)/pgdg-redhat-repo-latest.noarch.rpm && \
    dnf -qy module disable postgresql && \
    dnf install -y postgresql14-server

COPY scripts/* /scripts/
CMD /scripts/start.sh
