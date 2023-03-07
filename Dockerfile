# syntax=docker/dockerfile:1

FROM golang:1.20-alpine

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.15-r0 ffmpeg=5.1.2-r1 python3=3.10.10-r0

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2023.03.04/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ \
  && chmod +x /app/yt-dlp

CMD ["sh", "-c", "./hypebot -t=${TOKEN}"]
