.PHONY: all build test

all: build test

build:
	go build -o bin/pomo cmd/pomo/main.go

test:
	go test ./...