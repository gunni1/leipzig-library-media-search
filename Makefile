.PHONY: build clean

build-web:
	go build ./...
	go build -o bin/web cmd/web/main.go

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

