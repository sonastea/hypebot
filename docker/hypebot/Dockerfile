# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

ARG POToken
ARG PROXY_URL
ENV PATH="${PATH}:/app/"
ENV POToken=$POToken
ENV PROXY_URL=$PROXY_URL

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.37-r0

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2025.07.21/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp

FROM alpine:3.22.1

ARG POToken
ARG PROXY_URL
ENV PATH="${PATH}:/app/"
ENV POToken=$POToken
ENV PROXY_URL=$PROXY_URL

RUN apk add --no-cache ffmpeg=6.1.2-r2 python3=3.12.11-r0

COPY --from=builder /app/hypebot /app/hypebot
COPY --from=builder /app/yt-dlp /app/yt-dlp

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot --t=$TOKEN --discmds=$DISABLED_COMMANDS"]
