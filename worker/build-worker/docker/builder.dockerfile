FROM golang:1.8.1 as builder

MAINTAINER marwan.sameer@gmail.com

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad

COPY . /go/src/github.com/marwan-at-work/baghdad

RUN cd /go/src/github.com/marwan-at-work/baghdad/worker/build-worker && CGO_ENABLED=0 go build -a -ldflags '-s'

FROM busybox

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad/worker/build-worker

WORKDIR /go/src/github.com/marwan-at-work/baghdad/worker/build-worker

COPY --from=builder /go/src/github.com/marwan-at-work/baghdad/worker/build-worker/build-worker /go/src/github.com/marwan-at-work/baghdad/worker/build-worker

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs

ENTRYPOINT ["./build-worker"]
