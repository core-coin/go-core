# Build Gocore in a stock Go builder container
FROM golang:1.15-buster as builder
ARG ALLTOOLS

ADD . /go-core
RUN if [ -n "$ALLTOOLS" ] ; then cd /go-core && make all ; else cd /go-core && make gocore ; fi

# Pull Gocore into a second stage deploy alpine container
FROM buildpack-deps:buster-scm
ENV NETWORK=devin
ENV SYNCMODE=full
ENV GCMODE=full
ENV DATADIR=~/.core
ENV KEYDIR=$DATADIR/keystore
# Ports:
# 8545/tcp = HTTP-RPC
# 8546/tcp = WS
# 8547/tcp = GraphQL
# 30300/tcp 30300/udp = Peers

COPY --from=builder /go-core/build/bin/* /usr/local/bin/

EXPOSE 30300 30300/udp
CMD gocore --datadir=${DATADIR} --keystore=${KEYDIR} --${NETWORK} --syncmode=${SYNCMODE} --gcmode=${GCMODE}
