FROM rockylinux:8

ENV ETCD_VER=v3.5.2
ENV ETCD_DOWNLOAD_URL=https://github.com/etcd-io/etcd/releases/download
ENV STOLON_VER=v0.17.0
ENV STOLON_DOWNLOAD_URL=https://github.com/sorintlab/stolon/releases/download

RUN yum update -y && \
    yum install -y bind-utils make git iproute

# Install etcd
RUN cd $(mktemp -d) && \
    curl -L ${ETCD_DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o etcd-${ETCD_VER}-linux-amd64.tar.gz && \
    tar xzvf etcd-${ETCD_VER}-linux-amd64.tar.gz && \
    cd etcd-${ETCD_VER}-linux-amd64 && \
    mv etcd etcdctl etcdutl /usr/local/bin && \
    useradd -mrb /var/lib etcd

# Install stolon
RUN cd $(mktemp -d) && \
    curl -L ${STOLON_DOWNLOAD_URL}/${STOLON_VER}/stolon-${STOLON_VER}-linux-amd64.tar.gz -o stolon-${STOLON_VER}-linux-amd64.tar.gz && \
    tar xzvf stolon-${STOLON_VER}-linux-amd64.tar.gz && \
    mv stolon-${STOLON_VER}-linux-amd64/bin/* /usr/local/bin

RUN dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-x86_64/pgdg-redhat-repo-latest.noarch.rpm && \
    dnf -qy module disable postgresql && \
    dnf install -y postgresql14-server

COPY scripts/* /scripts/
CMD /scripts/start.sh
