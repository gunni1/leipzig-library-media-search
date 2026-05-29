build-web:
	go build ./...
	go test ./...
	go build -o bin/web main.go

build-amd64:
	go build ./...
	GOOS=linux GOARCH=amd64 go build -o bin/llms-amd64-linux main.go


build-arm64-linux:
	go build ./...
	GOOS=linux GOARCH=arm64 go build -o bin/llms-arm64-linux main.go

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

docker-build:
	docker build -t gunni1/leipzig-library-media-search:latest .
	docker push gunni1/leipzig-library-media-search:latest
