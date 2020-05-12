# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gcore android ios gcore-cross cvm all test clean
.PHONY: gcore-linux gcore-linux-386 gcore-linux-amd64 gcore-linux-mips64 gcore-linux-mips64le
.PHONY: gcore-linux-arm gcore-linux-arm-5 gcore-linux-arm-6 gcore-linux-arm-7 gcore-linux-arm64
.PHONY: gcore-darwin gcore-darwin-386 gcore-darwin-amd64
.PHONY: gcore-windows gcore-windows-386 gcore-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

gcore:
	$(GORUN) build/ci.go install ./cmd/gcore
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gcore\" to launch gcore."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gcore.aar\" to use the library."

ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gcore.framework\" to use the library."

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

gcore-cross: gcore-linux gcore-darwin gcore-windows gcore-android gcore-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gcore-*

gcore-linux: gcore-linux-386 gcore-linux-amd64 gcore-linux-arm gcore-linux-mips64 gcore-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-*

gcore-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gcore
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep 386

gcore-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gcore
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep amd64

gcore-linux-arm: gcore-linux-arm-5 gcore-linux-arm-6 gcore-linux-arm-7 gcore-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep arm

gcore-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gcore
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep arm-5

gcore-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gcore
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep arm-6

gcore-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gcore
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep arm-7

gcore-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gcore
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep arm64

gcore-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gcore
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep mips

gcore-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gcore
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep mipsle

gcore-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gcore
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep mips64

gcore-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gcore
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gcore-linux-* | grep mips64le

gcore-darwin: gcore-darwin-386 gcore-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gcore-darwin-*

gcore-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gcore
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-darwin-* | grep 386

gcore-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gcore
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-darwin-* | grep amd64

gcore-windows: gcore-windows-386 gcore-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gcore-windows-*

gcore-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gcore
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-windows-* | grep 386

gcore-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gcore
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gcore-windows-* | grep amd64
