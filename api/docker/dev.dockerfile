FROM golang:1.8.3

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad/api && \
    go get -u github.com/marwan-at-work/gowatch

WORKDIR /go/src/github.com/marwan-at-work/baghdad/api

CMD gowatch
