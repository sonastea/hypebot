# syntax=docker/dockerfile:1

FROM golang:1.18-alpine

# need gcc for go-sqlite3
RUN apk add build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build ./cmd/hypebot/

CMD ./hypebot -t=$TOKEN -g=$GUILD_ID
