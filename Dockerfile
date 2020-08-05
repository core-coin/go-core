# Build Gcore in a stock Go builder container
FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-core
RUN cd /go-core && make gcore
RUN cd /go-core && git clone http://github.com/core-coin/core-genesis

# Pull Gcore into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-core/build/bin/gcore /usr/local/bin/
COPY --from=builder /go-core/core-genesis/genesis.json /

EXPOSE 8545 8546 8547 30300 30300/udp
RUN gcore --datadir=/testdata init genesis.json
ENTRYPOINT ["gcore", "--datadir=/testdata", "--networkid", "3"]
