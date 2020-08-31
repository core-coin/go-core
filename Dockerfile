# Build Gocore in a stock Go builder container
FROM golang:1.13-alpine as builder
ARG ALLTOOLS
RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-core
RUN if [[ -n "$ALLTOOLS" ]] ; then \
      cd /go-core && make all \
    else \
      cd /go-core && make gocore \
    fi

# Pull Gocore into a second stage deploy alpine container
FROM alpine:latest
ENV NETWORK=mainnet
ENV SYNCMODE=fast
ENV GCMODE=full
ENV DATADIR=~/.core
ENV KEYDIR=$DATADIR/keystore
# Ports:
# 8545/tcp = HTTP-RPC
# 8546/tcp = WS
# 8547/tcp = GraphQL
# 30300/tcp 30300/udp = Peers

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-core/build/bin/* /usr/local/bin/

EXPOSE 30300 30300/udp
ENTRYPOINT ["gocore", "--datadir=${DATADIR}", "--keystore=${KEYDIR}", "--${NETWORK}", "--nat", "auto", "--syncmode=${SYNCMODE}", "--gcmode=${GCMODE}"]
