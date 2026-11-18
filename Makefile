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

.PHONY: test test-go test-images test-hypebot-image test-hypeserver-image
test: test-go test-images

test-go:
	@command -v ffmpeg >/dev/null || { echo "ffmpeg is required for the Go tests" >&2; exit 1; }
	go test -mod=readonly ./... -count=1 -timeout=20s

test-images: test-hypebot-image test-hypeserver-image

# Pull current base images and refresh apk layers even on a warm builder.
test-hypebot-image:
	docker buildx build --load --pull --no-cache-filter build-deps,runtime-deps -f docker/hypebot/Dockerfile -t hypebot:ci-test .
	docker run --rm --network none --entrypoint /bin/sh hypebot:ci-test -ec '\
		ldd ./hypebot; \
		ffmpeg -v error -f lavfi -i sine=duration=0.1 -af loudnorm -c:a libopus -f opus - > /dev/null; \
		python3 --version; deno --version; yt-dlp --version'

test-hypeserver-image:
	docker buildx build --load --pull --no-cache-filter build-deps -f docker/hypeserver/Dockerfile -t hypeserver:ci-test .
	docker run --rm --network none --entrypoint /bin/sh hypeserver:ci-test -ec 'ldd ./hypeserver'

clean:
	rm -f hypeserver hypebot

.PHONY: tidy
tidy:
	gofumpt -l -d .
	go mod tidy -v
