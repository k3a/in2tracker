.PHONY: test

PACKAGES = $(shell go list ./... | grep -v /vendor/)

gen:
	go generate $(PACKAGES)

test:
	go test -cover $(PACKAGES)

