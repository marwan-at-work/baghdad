FROM golang:1.8.1

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad/worker/deploy-sync && \
    go get -u github.com/marwan-at-work/gowatch

WORKDIR /go/src/github.com/marwan-at-work/baghdad/worker/deploy-sync

CMD gowatch
