# Stage Build
FROM golang:1.19 as build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd/server/main.go ./
COPY pkg ./pkg
RUN go mod download
RUN go build -o lib-api

# Stage Run
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/lib-api .
RUN chmod +x lib-api
CMD ["./lib-api"]
EXPOSE 8080