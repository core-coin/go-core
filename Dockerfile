FROM golang:1.15-buster as builder
ARG ALLTOOLS

ADD . /go-core
RUN if [ -n "$ALLTOOLS" ] ; then cd /go-core && make all ; else cd /go-core && make gocore ; fi

# Pull Gocore into a second stage deploy alpine container
FROM buildpack-deps:buster-scm

COPY --from=builder /go-core/build/bin/* /usr/local/bin/

EXPOSE 30300 30300/udp
