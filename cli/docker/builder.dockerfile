FROM golang:1.8.3

MAINTAINER marwan.sameer@gmail.com

RUN mkdir -p /go/src/github.com/marwan-at-work/baghdad

COPY . /go/src/github.com/marwan-at-work/baghdad

WORKDIR /go/src/github.com/marwan-at-work/baghdad/cli

RUN cd /go/src/github.com/marwan-at-work/baghdad/cli && CGO_ENABLED=0 go build -a -ldflags '-s' && \
    mv /go/src/github.com/marwan-at-work/baghdad/cli/cli /go/src/github.com/marwan-at-work/baghdad/cli/baghdad

CMD ["./baghdad"]
