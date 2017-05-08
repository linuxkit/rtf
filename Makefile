default: _build/rtf

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
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

_build:
	mkdir -p $@

_build/rtf: $(DEPS) lint test | _build
	@docker build -t rtf-dev -f Dockerfile.build .
	@docker run --rm \
		-e GOOS=${GOOS} -e GOARCCH=${GOARCH} \
		-v ${CURDIR}/_build:/go/src/github.com/linuxkit/rtf/_build \
		rtf-dev \
		make build

_build/rtf-darwin-amd64: $(DEPS) | _build
	@docker build -t rtf-dev -f Dockerfile.build .
	@docker run --rm \
		-e GOOS=darwin -e GOARCCH=amd64 \
		-v ${CURDIR}/_build:/go/src/github.com/linuxkit/rtf/_build \
		rtf-dev \
		make build-darwin-amd64

_build/rtf-windows-amd64: $(DEPS) lint test | _build
	@docker build -t rtf-dev -f Dockerfile.build .
	@docker run --rm \
		-e GOOS=windows -e GOARCCH=amd64 \
		-v ${CURDIR}/_build:/go/src/github.com/linuxkit/rtf/_build \
		rtf-dev \
		make build-windows-amd64

_build/rtf-linux-amd64: $(DEPS) lint test | _build
	@docker build -t rtf-dev -f Dockerfile.build .
	@docker run --rm \
		-e GOOS=linux -e GOARCCH=amd64 \
		-v ${CURDIR}/_build:/go/src/github.com/linuxkit/rtf/_build \
		rtf-dev \
		make build-linux-amd64

.PHONY: lint
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

.PHONY: build
build:
	@go build --ldflags "-X $(CMD_PKG).GitCommit=$(GIT_COMMIT) -X $(CMD_PKG).Version=$(VERSION)" -o _build/rtf github.com/linuxkit/rtf

.PHONY: build-%
build-%:
	@go build --ldflags "-X $(CMD_PKG).GitCommit=$(GIT_COMMIT) -X $(CMD_PKG).Version=$(VERSION)" -o _build/rtf-$* github.com/linuxkit/rtf

.PHONY: cross
cross: _build/rtf-darwin-amd64 _build/rtf-windows-amd64 _build/rtf-linux-amd64

.PHONY: test
test:
	@go test $(PKGS) 

.PHONY: install
install: rtf
	@cp -a $^ $(PREFIX)/bin/

.PHONY: clean
clean:
	@rm -rf rtf _build
