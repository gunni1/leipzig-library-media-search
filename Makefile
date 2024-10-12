build-web:
	go build ./...
	go test ./...
	go build -o bin/web main.go

build-amd64:
	go build ./...
	GOOS=linux GOARCH=amd64 go build -o bin/web main.go

build-arm:
	go build ./...
	GOOS=linux GOARCH=arm go build -o bin/web main.go

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

docker-build:
	docker build -t gunni1/lib-api:local .

docker-run:
	docker run -p 8080:8080 --rm --name lib-api gunni1/lib-api:local