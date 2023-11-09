# syntax=docker/dockerfile:1

FROM golang:1.21-alpine as builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.15-r5

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2023.10.13/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp

FROM alpine:3.17.5

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache ffmpeg=5.1.3-r0 python3=3.10.13-r0

COPY --from=builder /app/hypebot /app/hypebot
COPY --from=builder /app/yt-dlp /app/yt-dlp

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot -t=$TOKEN"]

