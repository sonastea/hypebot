# syntax=docker/dockerfile:1

FROM golang:1.18-alpine

# need gcc for go-sqlite3
RUN apk add --no-cache build-base=0.5-r3 bash=5.1.16-r2 ffmpeg=5.0.1-r1

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build ./cmd/hypebot/

CMD ["sh", "-c", "./hypebot -t=${TOKEN} -g=${GUILD_ID}"]
