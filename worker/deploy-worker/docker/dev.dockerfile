FROM golang:1.8.1

RUN mkdir -p /go/src/github.com/workco/marwan-at-work/baghdad/deploy-worker && \
    go get -u github.com/marwan-at-work/gowatch && \
    curl -sSL https://get.docker.com/ | sh

WORKDIR /go/src/github.com/workco/marwan-at-work/baghdad/deploy-worker

CMD gowatch
