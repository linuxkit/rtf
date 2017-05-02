.PHONY: build 

BINARIES:=rt-local
PACKAGES:=$(shell go list ./... | grep -v vendor)

build: $(BINARIES)

clean:
	rm -rf $(BINARIES)

rt-local:
	@go build -o rt-local main.go

test: 
	@go test $(PACKAGES)
