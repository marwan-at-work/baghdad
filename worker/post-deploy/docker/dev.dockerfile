FROM golang:1.8.3

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad/worker/post-deploy && \
    go get -u github.com/marwan-at-work/gowatch

WORKDIR /go/src/github.com/marwan-at-work/baghdad/worker/post-deploy

CMD gowatch
