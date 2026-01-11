.DEFAULT_GOAL := build

BINARY := urlshortener-backend

.PHONY: fmt vet build clean test

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o $(BINARY) ./cmd

test:
	go test ./...

clean:
	rm -f $(BINARY)
