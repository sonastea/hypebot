# syntax=docker/dockerfile:1

FROM golang:1.21-alpine

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.15-r5 ffmpeg=6.0-r15 python3=3.11.5-r0

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2023.07.06/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp

CMD ["sh", "-c", "./hypebot -t=$TOKEN"]
