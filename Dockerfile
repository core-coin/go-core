FROM golang:alpine as builder
RUN apk add make gcc g++ musl-dev linux-headers
ADD . /go-core
RUN cd /go-core && make all

FROM alpine:3.14
RUN apk add gcc g++
COPY --from=builder /go-core/build/bin/* /usr/local/bin/
EXPOSE 30300 30300/udp
