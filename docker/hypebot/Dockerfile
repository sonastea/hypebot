# syntax=docker/dockerfile:1

FROM golang:1.20.2-alpine as builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.15-r0

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2023.03.04/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp

FROM alpine:3.17.3

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache ffmpeg=5.1.3-r0 python3=3.10.11-r0

COPY --from=builder /app/ /app/

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot -t=$TOKEN"]

