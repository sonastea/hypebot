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
3. Run the binary with `./hypebot[.exe] -t "bot token" -g "guild id"` depending on your os.

## Running as docker container

Building the image

1. Go to root of the folder after cloning this repository.
2. Build the image with `docker build -t hypebot .`

Run the image in a container

1. Create an .env file with these three variables: 
   - `TOKEN`    - Your discord bot token.
   - `GUILD_ID` - Your discord server id.
   - `POToken`  - Proof of Origin token. *(Optional)*
   - `PROXY_URL` – A proxy url to use when making requests. *(Optional)*
   - `CUSTOM_STATUS` – The custom status message for your bot. *(Optional)*

   <br>

   **If using `POToken`, follow steps 2 and 3 below to set it up properly.**

2. Obtain your POToken. [*(Optional)*](https://github.com/yt-dlp/yt-dlp/wiki/Extractors#manually-acquiring-a-po-token-from-a-browser-for-use-when-logged-in)
3. Export your cookies as a `cookies.txt` in the root directory. [*(Optional)*](https://github.com/yt-dlp/yt-dlp/issues/10927#issuecomment-2332602512)
4. Use your discord bot token for "TOKEN" and your discord server's id for "GUILD_ID".
5. Create a docker container from the hypebot image with `docker run --env-file .env -d hypebot`.
6. Docker hypebot container should be running in the background.
   > Name your container with `docker run --env-file .env -d --name <container_name> hypebot`.

## Contact Me

Message me on Discord `nastea` if you have any questions. Feel free to report any bugs or create a pull request, and I'll try to respond as soon as I can.
Click the link [here](https://github.com/sonastea/hypebot/issues/new) to create an issue.
