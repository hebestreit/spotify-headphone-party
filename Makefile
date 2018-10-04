MAKEFLAGS += --warn-undefined-variables
BINARY ?= spotify-headphone-party

.PHONY: %

all: clean deps build

deps:
	dep ensure

build:
	CGO_ENABLED=0 go build -o bin/$(BINARY) -v -i --ldflags=--s main.go

clean:
	rm -rf bin/$(BINARY) vendor/*