discmds ?= ""

all: \
	build-bot build-server

build-bot:
	go build ./cmd/hypebot

run-bot: build-bot
	./hypebot --t=$(TOKEN) --bid=$(BOT_ID) --g=$(GUILD_ID) --discmds=$(discmds)

build-server:
	go build ./cmd/hypeserver

run-server: build-server
	./hypeserver

clean:
	rm -f hypeserver hypebot

.PHONY: tidy
tidy:
	gofumpt -l -d .
	go mod tidy -v
