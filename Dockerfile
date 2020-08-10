# Build Gocore in a stock Go builder container
FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-core
RUN cd /go-core && make gocore
RUN cd /go-core && git clone http://github.com/core-coin/core-genesis

# Pull Gocore into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-core/build/bin/gocore /usr/local/bin/
COPY --from=builder /go-core/core-genesis/genesis.json /

EXPOSE 30300 30300/udp
RUN gocore --datadir=/testdata init genesis.json
ENTRYPOINT ["gocore", "--datadir=/testdata", "--networkid", "3"]
