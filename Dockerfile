# Build Gocore in a stock Go builder container
FROM golang:1.15-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /go-core
RUN cd /go-core && make gocore

# Pull Gocore into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-core/build/bin/gocore /usr/local/bin/

EXPOSE 8545 8546 30300 30300/udp
ENTRYPOINT ["gocore"]
