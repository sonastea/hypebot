# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.26-r0

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2024.11.04/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp

FROM alpine:3.20

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache ffmpeg=6.1.1-r8 python3=3.12.7-r0

COPY --from=builder /app/hypebot /app/hypebot
COPY --from=builder /app/yt-dlp /app/yt-dlp

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot --t=$TOKEN --discmds=$DISABLED_COMMANDS"]
