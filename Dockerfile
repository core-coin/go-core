# Build Gocore in a stock Go builder container
FROM golang:1.13-alpine as builder
ARG NETWORK=mainnet
ARG DATADIR=~/.core
ARG ALLTOOLS
ARG EXPOSEPORTS=30300 30300/udp
# Exposing ports:
# 8545:8545/tcp = HTTP-RPC
# 8546:8546/tcp = WS
# 8547:8547/tcp = GraphQL
# 30300 30300/udp = Peers

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-core
RUN if [[ -n "$ALLTOOLS" ]] ; then \
      cd /go-core && make all \
    else \
      cd /go-core && make gocore \
    fi

# Pull Gocore into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-core/build/bin/gocore /usr/local/bin/

EXPOSE $EXPOSEPORTS
ENTRYPOINT ["gocore", "--datadir=${DATADIR}", "--${NETWORK}", "--nat", "auto"]
