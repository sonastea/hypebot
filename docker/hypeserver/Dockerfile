# syntax=docker/dockerfile:1

FROM golang:1.20.2-alpine as builder

ENV PATH="${PATH}:/app/"

RUN apk add --no-cache build-base=0.5-r3 bash=5.2.15-r0

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build ./cmd/hypeserver/

FROM alpine:3.17.3

ENV PATH="${PATH}:/app/"

COPY --from=builder /app/ /app/

WORKDIR /app

CMD ["sh", "-c", "./hypeserver"]