MAKEFLAGS += --warn-undefined-variables
IMAGE ?= hebestreit/spotify-party
BINARY ?= spotify-party

.PHONY: %

all: clean deps build docker-build

deps:
	dep ensure

build:
	CGO_ENABLED=0 go build -o bin/$(BINARY) -v -i --ldflags=--s main.go

docker-build:
	docker build -t $(IMAGE):latest .

clean:
	rm -rf bin/$(BINARY) vendor/*