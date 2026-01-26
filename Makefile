VERSION ?= $(shell git describe --tags --always || echo "dev")

MAIN_PATH = ./cmd/pomo/pomo.go
BINARY_NAME = pomo

.PHONY: build install clean

build:
	go build -o $(BINARY_NAME) -ldflags="-X main.version=$(VERSION)" $(MAIN_PATH)

install:
	go install -ldflags="-X main.version=$(VERSION)" $(MAIN_PATH)

clean:
	rm -f $(BINARY_NAME)