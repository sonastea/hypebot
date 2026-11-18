# syntax=docker/dockerfile:1

FROM golang:1.27-alpine3.23 AS build-deps

RUN apk add --no-cache build-base bash

FROM build-deps AS builder

ENV PATH="${PATH}:/app/"

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

# Let parallel image builds reuse compilation of shared CGO dependencies.
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go build -mod=readonly ./cmd/hypeserver/

FROM alpine:3.23

ENV PATH="${PATH}:/app/"

COPY --from=builder /app/hypeserver /app/hypeserver

WORKDIR /app

CMD ["sh", "-c", "./hypeserver"]
