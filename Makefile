VERSION ?= $(shell git describe --tags --always || echo "dev")

MAIN_PKG = ./cmd/pomo
BINARY_NAME = pomo

WEB_DIR = web
UI_DIST = internal/webui/dist

.PHONY: build install clean test lint web-install web-dev web-build ui

build:
	go build -o $(BINARY_NAME) -ldflags="-X main.version=$(VERSION)" $(MAIN_PKG)

# Svelte dashboard, served by the daemon (embedded via go:embed).
web-install:
	cd $(WEB_DIR) && npm install

web-dev:
	cd $(WEB_DIR) && npm run dev

web-build:
	cd $(WEB_DIR) && npm run build

# Build the UI and sync it into the Go embed dir. Run before `make build` to
# bake the latest dashboard into the binary.
ui: web-build
	rm -rf $(UI_DIST)
	mkdir -p $(UI_DIST)
	cp -r $(WEB_DIR)/build/. $(UI_DIST)/
	touch $(UI_DIST)/.gitkeep

install:
	go install -ldflags="-X main.version=$(VERSION)" $(MAIN_PKG)

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)

lint:
	golangci-lint run ./...
