# Stage Build
FROM golang:1.20 as build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY . ./

RUN go mod download
RUN go build -o web

# Stage Run
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/web .
RUN chmod +x web
CMD ["./web"]
EXPOSE 8080