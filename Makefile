
VERSION := $(shell git describe --always)

all: build test

build:
	go build -o pomo -ldflags="-X main.version=$(VERSION)" cmd/pomo/main.go

test:
	go test ./...
