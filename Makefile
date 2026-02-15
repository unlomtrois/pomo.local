VERSION ?= $(shell git describe --tags --always || echo "dev")

MAIN_PKG = ./cmd/pomo
BINARY_NAME = pomo

.PHONY: build install clean test

install:
	go install -ldflags="-X main.version=$(VERSION)" $(MAIN_PKG)

build:
	go build -o $(BINARY_NAME) -ldflags="-X main.version=$(VERSION)" $(MAIN_PKG)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
