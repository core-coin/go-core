.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/core-coin/go-core/cmd/swarm
	go build -o bin/geth     github.com/core-coin/go-core/cmd/geth
	go build -o bin/bootnode github.com/core-coin/go-core/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/core-coin/go-core/swarm/...
