# BUILD
FROM golang:1.8.1 as builder

MAINTAINER marwan.sameer@gmail.com

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad

COPY . /go/src/github.com/marwan-at-work/baghdad

WORKDIR /go/src/github.com/marwan-at-work/baghdad/api

RUN cd /go/src/github.com/marwan-at-work/baghdad/api && CGO_ENABLED=0 go build -a -ldflags '-s'

# RUN
FROM busybox

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad/api

WORKDIR /go/src/github.com/marwan-at-work/baghdad/api

COPY --from=builder /go/src/github.com/marwan-at-work/baghdad/api/api /go/src/github.com/marwan-at-work/baghdad/api

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["./api"]
