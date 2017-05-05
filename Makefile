.PHONY: build 

VERSION="0.0" # dummy for now
GIT_COMMIT=$(shell git rev-list -1 HEAD)
PACKAGES:=$(shell go list ./... | grep -v vendor)
GOLINT:=$(shell command -v golint)
DEPS=$(wildcard *.go) Makefile

CMD_PKG=github.com/linuxkit/rtf/cmd

rt-local: $(DEPS)
	go build --ldflags "-X $(CMD_PKG).GitCommit=$(GIT_COMMIT) -X $(CMD_PKG).Version=$(VERSION)" -o $@ main.go

.PHONY: install
install:
	@go install .

clean:
	@rm -rf rt-local

test: 
	@go test $(PACKAGES)

lint:
ifndef $(GOLINT)
	error("Please install golint! go get -u github.com/tool/lint")
endif
