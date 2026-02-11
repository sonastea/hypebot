# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.3.3-r1

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypeserver/

FROM alpine:3.23.0

ENV PATH="${PATH}:/app/"

COPY --from=builder /app/hypeserver /app/hypeserver

WORKDIR /app

CMD ["sh", "-c", "./hypeserver"]
