# Build Gcore in a stock Go builder container
FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-ethereum
RUN cd /go-ethereum && make gcore

# Pull Gcore into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-ethereum/build/bin/gcore /usr/local/bin/

EXPOSE 8545 8546 8547 30300 30300/udp
ENTRYPOINT ["gcore"]
