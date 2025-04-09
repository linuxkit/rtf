VERSION="0.0" # dummy for now
GIT_COMMIT=$(shell git rev-list -1 HEAD)
CMD_PKG=github.com/linuxkit/rtf/cmd
PKGS:=$(shell go list ./... | grep -v vendor)

GOOS?=$(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH?=amd64
ifneq ($(GOOS),linux)
CROSS+=-e GOOS=$(GOOS)
endif
ifneq ($(GOARCH),amd64)
CROSS+=-e GOARCH=$(GOARCH)
endif

DEPS=Makefile main.go
DEPS+=$(wildcard cmd/*.go)
DEPS+=$(wildcard local/*.go)
DEPS+=$(wildcard logger/*.go)
DEPS+=$(wildcard sysinfo/*.go)

PREFIX?=/usr/local

GOLINT:=$(shell command -v golangci-lint 2> /dev/null)
INEFFASSIGN:=$(shell command -v ineffassign 2> /dev/null)

default: rtf

# Build with docker
GO_COMPILE=linuxkit/go-compile:7cac05c5588b3dd6a7f7bdb34fc1da90257394c7
.PHONY: build-with-docker
build-with-docker: tmp_rtf_bin.tar
	tar xf $<
	rm $<

tmp_rtf_bin.tar: $(DEPS)
	tar cf - . | docker run --rm --net=none --log-driver=none -i $(CROSS) $(GO_COMPILE) --package github.com/linuxkit/rtf --ldflags "-X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)" -o rtf > $@


# Build local (default)
rtf: $(DEPS)
	go build --ldflags "-X $(CMD_PKG).GitCommit=$(GIT_COMMIT) -X $(CMD_PKG).Version=$(VERSION)" -o $@

.PHONY: lint
lint:
ifndef GOLINT
	$(error "Please install golangci-lint! http://golangci-lint.run")
endif
ifndef INEFFASSIGN
	$(error "Please install ineffassign! go install github.com/gordonklaus/ineffassign@latest")
endif
	@echo "+ $@: golangci-lint, gofmt, go vet, ineffassign"
	# golangci-lint
	golangci-lint run ./...
	# gofmt
	@test -z "$$(gofmt -s -l .| grep -v .pb. | grep -v vendor/ | tee /dev/stderr)"
ifeq ($(GOOS),)
	# govet
	@test -z "$$(go tool vet -printf=false . 2>&1 | grep -v vendor/ | tee /dev/stderr)"
endif
	# ineffassign
	@test -z $$(ineffassign ./... | tee /dev/stderr)

.PHONY: install-deps
install-deps:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2
	go install github.com/gordonklaus/ineffassign@latest

.PHONY: test
test: rtf lint
	@go test $(PKGS)

.PHONY: install
install: rtf
	cp -a $^ $(PREFIX)/bin/

.PHONY: clean
clean:
	rm -f rtf
