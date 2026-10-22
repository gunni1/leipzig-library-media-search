# Stage Build
FROM golang:1.26 AS build

ARG TARGETARCH=arm64
ENV GOOS=linux
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o bin/web main.go && mkdir -p data

# Stage Run
FROM --platform=linux/arm64 gcr.io/distroless/static-debian13
WORKDIR /
COPY --from=build /app/bin/web /web
# Create writable data directory owned by nonroot (UID/GID 65532)
COPY --from=build --chown=65532:65532 /app/data/ /data/
VOLUME /data
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/web", "-data-dir", "/data"]
