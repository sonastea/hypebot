# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

ARG POToken
ARG PROXY_URL
ENV PATH="${PATH}:/app/"
ENV POToken=$POToken
ENV PROXY_URL=$PROXY_URL

RUN apk add --no-cache build-base=0.5-r3 bash=5.3.3-r1 unzip=6.0-r16

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2025.12.08/yt-dlp /app/yt-dlp
ADD https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip /tmp/deno.zip

RUN unzip /tmp/deno.zip -d /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypebot/ && chmod +x /app/yt-dlp && chmod +x /app/deno

FROM alpine:3.23.0

ARG POToken
ARG PROXY_URL
ENV PATH="${PATH}:/app/"
ENV POToken=$POToken
ENV PROXY_URL=$PROXY_URL

RUN apk add --no-cache ffmpeg=8.0.1-r0 python3=3.12.12-r0

COPY --from=builder /app/hypebot /app/hypebot
COPY --from=builder /app/yt-dlp /app/yt-dlp
COPY --from=builder /app/deno /app/deno

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot --t=$TOKEN --discmds=$DISABLED_COMMANDS"]
