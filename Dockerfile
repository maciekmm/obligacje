FROM golang:1.26-trixie AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /obligacje ./cmd/main.go


FROM debian:trixie-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates libreoffice-calc && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /obligacje /usr/local/bin/obligacje

VOLUME ["/data"]
WORKDIR /data

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/obligacje"]
