build-web:
	go build ./...
	go test ./...
	go build -o bin/web main.go

docker:
	docker build -t gunni1/lib-api:local .

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

