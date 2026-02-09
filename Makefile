.PHONY: build test lint run clean install

BINARY := aicrawler
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) ./cmd/aicrawler

test:
	go test ./... -v

lint:
	go vet ./...
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
	go clean -testcache

install: build
	cp $(BINARY) $(GOPATH)/bin/ 2>/dev/null || cp $(BINARY) ~/go/bin/

tidy:
	go mod tidy
