default: rtf

DEPS=Makefile main.go
DEPS+=$(wildcard cmd/*.go)
DEPS+=$(wildcard local/*.go)
DEPS+=$(wildcard logger/*.go)
DEPS+=$(wildcard sysinfo/*.go)

GOLINT:=$(shell command -v golint 2> /dev/null)
INEFFASSIGN:=$(shell command -v ineffassign 2> /dev/null)

PREFIX?=/usr/local
VERSION="0.0" # dummy for now
GIT_COMMIT=$(shell git rev-list -1 HEAD)
CMD_PKG=github.com/linuxkit/rtf/cmd
PKGS:=$(shell go list ./... | grep -v vendor)

rtf: $(DEPS) lint test
	go build --ldflags "-X $(CMD_PKG).GitCommit=$(GIT_COMMIT) -X $(CMD_PKG).Version=$(VERSION)" -o $@ github.com/linuxkit/rtf

lint:
ifndef GOLINT
	$(error "Please install golint! go get -u github.com/tool/lint")
endif
ifndef INEFFASSIGN
	$(error "Please install ineffassign! go get -u github.com/gordonklaus/ineffassign")
endif
	@echo "+ $@: golint, gofmt, go vet, ineffassign"
	# golint
	@test -z "$(shell find . -type f -name "*.go" -not -path "./vendor/*" -exec golint {} \; | tee /dev/stderr)"
	# gofmt
	@test -z "$$(gofmt -s -l .| grep -v .pb. | grep -v vendor/ | tee /dev/stderr)"
	# govet
	@test -z "$$(go tool vet -printf=false . 2>&1 | grep -v vendor/ | tee /dev/stderr)"
	# ineffassign
	@test -z $(find . -type f -name "*.go" -not -path "*/vendor/*" -exec ineffassign {} \; | tee /dev/stderr)

PHONY: test
test:
	go test $(PKGS) 

PHONY: install
install: rtf
	cp -a $^ $(PREFIX)/bin/

PHONY: clean
clean:
	rm -rf rtf
