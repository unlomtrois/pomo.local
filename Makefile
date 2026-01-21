.PHONY: all build test

all: build test

build:
	go build -o pomo cmd/pomo/main.go

test:
	go test ./...