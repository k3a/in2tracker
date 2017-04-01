.PHONY: gen test build-deps

PACKAGES = $(shell go list ./... | grep -v /vendor/)

gen: build-deps
	go generate $(PACKAGES)

test:
	go test -cover $(PACKAGES)

build-deps:
	go get -u github.com/jteeuwen/go-bindata/...

