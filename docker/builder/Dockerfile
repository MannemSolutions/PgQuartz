FROM rockylinux:8
ENV GOVER=1.18.2

RUN dnf update -y && \
    dnf groupinstall -y "Development Tools" && \
    dnf install -y bind-utils make git iproute

# Install GoLang
RUN cd $(mktemp -d) && \
    curl -L https://go.dev/dl/go${GOVER}.linux-amd64.tar.gz -o go${GOVER}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && tar -C /usr/local -xzf go${GOVER}.linux-amd64.tar.gz && \
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/bashrc

COPY scripts/* /scripts/
CMD /scripts/start.sh
