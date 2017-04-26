.PHONY: build 

BINARIES:=rt-local
PACKAGES:=$(shell go list ./... | grep -v vendor)

build: $(BINARIES)

clean:
	rm -rf $(BINARIES)

rt-local:
	@go build rt-local.go

test: 
	@go test $(PACKAGES)
