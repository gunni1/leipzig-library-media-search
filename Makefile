.PHONY: build clean

build-server:
	go build ./...
	go build -o bin/server cmd/server/main.go

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

