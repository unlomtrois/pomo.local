VERSION ?= $(shell git describe --tags --always || echo "dev")

MAIN_PKG = ./cmd/pomo
BINARY_NAME = pomo

WEB_DIR = web

.PHONY: build install clean test lint web-install web-dev web-build

build:
	go build -o $(BINARY_NAME) -ldflags="-X main.version=$(VERSION)" $(MAIN_PKG)

# Svelte dashboard (served by the daemon on pomo.local once embed wiring lands).
web-install:
	cd $(WEB_DIR) && npm install

web-dev:
	cd $(WEB_DIR) && npm run dev

web-build:
	cd $(WEB_DIR) && npm run build

install:
	go install -ldflags="-X main.version=$(VERSION)" $(MAIN_PKG)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)

lint:
	golangci-lint run ./...
