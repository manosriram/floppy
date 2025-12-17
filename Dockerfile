# syntax=docker/dockerfile:1

FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Copy module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the binary (cmd/main.go => package main in ./cmd)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/floppy ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk update && apk add --no-cache ca-certificates ffmpeg

COPY --from=builder /out/floppy /app/floppy

# Runtime assets
COPY web /app/web
COPY static /app/static

# The app expects a file named "config" in the working directory.
# We'll mount it via docker-compose, but create a placeholder so the path exists.
RUN touch /app/config

# Generated assets directories (can be persisted via volumes)
RUN mkdir -p /app/.thumbs /app/.hls

EXPOSE 5050

ENTRYPOINT ["/app/floppy"]
