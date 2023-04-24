.PHONY: build clean


build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

