# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gocore android ios gocore-cross cvm all test clean
.PHONY: gocore-linux gocore-linux-386 gocore-linux-amd64 gocore-linux-mips64 gocore-linux-mips64le
.PHONY: gocore-linux-arm gocore-linux-arm-5 gocore-linux-arm-6 gocore-linux-arm-7 gocore-linux-arm64
.PHONY: gocore-darwin gocore-darwin-386 gocore-darwin-amd64
.PHONY: gocore-windows gocore-windows-386 gocore-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

gocore:
	$(GORUN) build/ci.go install ./cmd/gocore
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gocore\" to launch gocore."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gocore.aar\" to use the library."

ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gocore.framework\" to use the library."

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gocore-cross: gocore-linux gocore-darwin gocore-windows gocore-android gocore-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gocore-*

gocore-linux: gocore-linux-386 gocore-linux-amd64 gocore-linux-arm gocore-linux-mips64 gocore-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-*

gocore-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gocore
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep 386

gocore-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gocore
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep amd64

gocore-linux-arm: gocore-linux-arm-5 gocore-linux-arm-6 gocore-linux-arm-7 gocore-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep arm

gocore-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gocore
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep arm-5

gocore-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gocore
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep arm-6

gocore-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gocore
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep arm-7

gocore-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gocore
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep arm64

gocore-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gocore
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep mips

gocore-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gocore
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep mipsle

gocore-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gocore
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep mips64

gocore-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gocore
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gocore-linux-* | grep mips64le

gocore-darwin: gocore-darwin-386 gocore-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gocore-darwin-*

gocore-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gocore
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-darwin-* | grep 386

gocore-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gocore
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gocore-darwin-* | grep amd64

gocore-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CCX=x86_64-w64-mingw32-gcc $(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gocore
	@echo "Windows amd64 compilation done:"
	@ls -ld $(GOBIN)/gocore-windows-* | grep amd64
