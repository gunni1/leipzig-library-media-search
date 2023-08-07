.PHONY: build clean

build-bot:
	go build ./...
	go build -o bin/bot cmd/bot/main.go

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

