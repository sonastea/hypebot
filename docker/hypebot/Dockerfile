# syntax=docker/dockerfile:1

FROM golang:1.27-alpine3.23 AS build-deps

RUN apk add --no-cache build-base bash unzip

FROM build-deps AS builder

ARG POToken
ARG PROXY_URL
ENV PATH="${PATH}:/app/"
ENV POToken=$POToken
ENV PROXY_URL=$PROXY_URL

WORKDIR /app

ADD https://github.com/yt-dlp/yt-dlp/releases/download/2026.07.04/yt-dlp /app/yt-dlp

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

# Let parallel image builds reuse compilation of shared CGO dependencies.
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go build -mod=readonly ./cmd/hypebot/ && chmod +x /app/yt-dlp

FROM alpine:3.23 AS runtime-deps

RUN apk add --no-cache ffmpeg python3 deno

FROM runtime-deps

ARG POToken
ARG PROXY_URL
ENV PATH="${PATH}:/app/"
ENV POToken=$POToken
ENV PROXY_URL=$PROXY_URL

COPY --from=builder /app/hypebot /app/hypebot
COPY --from=builder /app/yt-dlp /app/yt-dlp

RUN mkdir -p /app/songs/

WORKDIR /app

CMD ["sh", "-c", "./hypebot --t=$TOKEN"]
