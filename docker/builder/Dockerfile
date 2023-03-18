FROM rockylinux:8
ENV GOVER=1.20.2

RUN dnf update -y && \
    dnf groupinstall -y "Development Tools" && \
    dnf install -y bind-utils make git iproute

COPY scripts/* /scripts/
RUN /scripts/install_go.sh
CMD /scripts/start.sh
