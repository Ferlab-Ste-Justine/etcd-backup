FROM golang:1.22-bullseye

ENV CGO_ENABLED=0

WORKDIR /opt
COPY . .

RUN go build .

RUN wget -O /opt/etcd-v3.5.17-linux-amd64.tar.gz https://storage.googleapis.com/etcd/v3.5.17/etcd-v3.5.17-linux-amd64.tar.gz && \
    mkdir -p /opt/etcd && \
    tar xzvf /opt/etcd-v3.5.17-linux-amd64.tar.gz -C /opt/etcd

FROM scratch

COPY --from=0 /opt/etcd-backup /bin/
COPY --from=0 /opt/etcd/etcd-v3.5.17-linux-amd64/etcdutl /bin/

ENV WORKING_DIR="/opt"

ENTRYPOINT ["/bin/etcd-backup"]