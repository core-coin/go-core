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

gocore-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CCX=x86_64-w64-mingw32-gcc $(GORUN) build/ci.go install ./cmd/gocore
	@echo "Done building."
	@echo "Run \"$(GOBIN)\gocore.exe\" to launch gocore."

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
