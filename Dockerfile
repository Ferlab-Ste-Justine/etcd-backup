FROM golang:1.22-bullseye

ENV CGO_ENABLED=0

WORKDIR /opt
COPY . .

RUN go build .

FROM scratch

COPY --from=0 /opt/etcd-backup /bin/

ENV WORKING_DIR="/opt"

ENTRYPOINT ["/bin/etcd-backup"]