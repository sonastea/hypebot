# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

ENV PATH="${PATH}:/app/"

# need gcc for go-sqlite3
RUN apk add --no-cache build-base=0.5-r3 bash=5.1.16-r2 ffmpeg=5.0.1-r1 curl=7.83.1-r4 python3=3.10.8-r0

WORKDIR /app

RUN curl https://github.com/yt-dlp/yt-dlp/releases/download/2023.02.17/yt-dlp
COPY yt-dlp ./
RUN chmod +x ./yt-dlp

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build ./cmd/hypebot/

CMD ["sh", "-c", "./hypebot -t=${TOKEN} -g=${GUILD_ID}"]
