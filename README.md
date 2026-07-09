# HypeBot

> HypeBot is a discord bot that plays a user selected song upon entering a voice channel.

## Running HypeBot

Download the latest binaries

1. Go to [releases](https://github.com/sonastea/hypebot/releases) and download the latest version.
2. Follow the provided instructions and you are set.

Building from source

1. Clone this repository.
2. Build the binary with `go build ./cmd/hypebot`
   and an executable named "hypebot" will be created.
3. Export your cookies as a `cookies.txt` in the root directory. [(Required)](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#exporting-youtube-cookies)
4. Run the binary with `./hypebot[.exe] -t "bot token" -g "guild id"` depending on your os.
   - Use `-disable_potoken=false` flag to enable POToken (requires `POToken` env var).

## Running as docker container

Building the image

1. Go to root of the folder after cloning this repository.
2. Build the image with `docker build -t hypebot .`

Run the image in a container

1. Create an .env file with these variables: 
   - `TOKEN`    - Your discord bot token.
   - `GUILD_ID` - Your discord server id.
   - `PROXY_URL` – A proxy url to use when making requests. *(Optional)*
   - `CUSTOM_STATUS` – The custom status message for your bot. *(Optional)*
   - `POToken`  - Proof of Origin token. Only used when `-disable_potoken=false`. *(Optional)*

2. Export your cookies as a `cookies.txt` in the root directory. [(Required)](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#exporting-youtube-cookies)
3. *(Optional)* If you want to use POToken, pass `-disable_potoken=false` flag and provide your `POToken` env var. [Learn more](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#po-token-guide)
4. Use your discord bot token for "TOKEN" and your discord server's id for "GUILD_ID".
5. Create a docker container from the hypebot image with `docker run --env-file .env -d hypebot`.
6. Docker hypebot container should be running in the background.
   > Name your container with `docker run --env-file .env -d --name <container_name> hypebot`.

## Running in development

Run both apps locally using the published Docker images with bind-mounted data.

Building the images locally

1. From the root of the repository, build HypeBot with `docker build . -t sonastea/hypebot:latest -f ./docker/hypebot/Dockerfile`.
2. Build HypeServer with `docker build . -t sonastea/hypeserver:latest -f ./docker/hypeserver/Dockerfile`.

HypeBot:

```bash
docker run -d --name hypebot \
              --mount type=bind,src="$(pwd)"/songs,target=/app/songs \
              --mount type=bind,src="$(pwd)"/hypebase.db,target=/app/hypebase.db \
              --mount type=bind,src="$(pwd)"/cookies.txt,target=/app/cookies.txt \
              sonastea/hypebot:latest hypebot -t=$TOKEN
```

HypeServer:

```bash
docker run -d --name hypeserver \
              -p 3000:3000 \
              --mount type=bind,src="$(pwd)"/hypebase.db,target=/app/hypebase.db \
              sonastea/hypeserver:latest
```

## Contact Me

Message me on Discord `nastea` if you have any questions. Feel free to report any bugs or create a pull request, and I'll try to respond as soon as I can.
Click the link [here](https://github.com/sonastea/hypebot/issues/new) to create an issue.
