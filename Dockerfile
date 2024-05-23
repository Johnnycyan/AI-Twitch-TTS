# Build stage
FROM golang:alpine AS build

# Install UPX and other dependencies
RUN apk add --no-cache upx

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY static/ ./static/

RUN go build -ldflags="-s -w" -o main .

# Compress the binary with UPX
RUN upx --brute main

# Final stage
FROM alpine:3.20

# Install ffmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app

# Copy the compressed binary from the build stage
COPY --from=build /app/main .

# Copy the necessary files
COPY static/ ./static/

EXPOSE 8080

CMD ["./main", "8080"]
