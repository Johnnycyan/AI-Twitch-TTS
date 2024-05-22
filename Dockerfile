FROM golang:alpine

# Install ffmpeg and other dependencies
RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY *.html ./
COPY static/ ./static/
COPY reverb.wav ./

RUN go build -o main .

EXPOSE 8080

CMD ["./main", "8080"]
