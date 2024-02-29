build-web:
	go build ./...
	go test ./...
	go build -o bin/web main.go

build-cli:
	go build ./...
	go build -o bin/cli cmd/cli/main.go

docker:
	docker build -t gunni1/lib-api:local .

docker-run:
	docker run -p 8080:8080 --rm --name lib-api gunni1/lib-api:local



