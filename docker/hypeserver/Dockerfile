# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.37-r0

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypeserver/

FROM alpine:3.21.3

ENV PATH="${PATH}:/app/"

COPY --from=builder /app/hypeserver /app/hypeserver

WORKDIR /app

CMD ["sh", "-c", "./hypeserver"]
