FROM golang:1.8.3 AS builder

MAINTAINER marwan.sameer@gmail.com

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad

COPY . /go/src/github.com/marwan-at-work/baghdad

RUN cd /go/src/github.com/marwan-at-work/baghdad/worker/docker-sweeper && CGO_ENABLED=0 go build -a -ldflags '-s'

FROM busybox

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad/worker/docker-sweeper

WORKDIR /go/src/github.com/marwan-at-work/baghdad/worker/docker-sweeper

COPY --from=builder /go/src/github.com/marwan-at-work/baghdad/worker/docker-sweeper/docker-sweeper /go/src/github.com/marwan-at-work/baghdad/worker/docker-sweeper

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["./docker-sweeper"]
