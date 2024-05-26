# syntax=docker/dockerfile:1

FROM golang:1.22-alpine as builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.26-r0

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2024.04.09/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp

FROM alpine:3.19

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache ffmpeg=6.1.1-r0 python3=3.11.9-r0

COPY --from=builder /app/hypebot /app/hypebot
COPY --from=builder /app/yt-dlp /app/yt-dlp

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot -t=$TOKEN"]
