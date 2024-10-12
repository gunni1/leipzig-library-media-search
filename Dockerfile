# Stage Build
FROM golang:1.23 as build

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY . ./
RUN go mod download

RUN make build-web

# Stage Run
FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=build /app/bin/web /web
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/web"]
