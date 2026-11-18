package hypebot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/robrotheram/dca"
	"github.com/sonastea/hypebot/internal/datastore/themesong"
)

const (
	SEND_TIMEOUT        = 5 * time.Second
	POST_PLAYBACK_DELAY = 1500 * time.Millisecond
)

type VideoMetaData struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	OriginalURL string `json:"original_url"`
	UploadDate  string `json:"upload_date"`
	Duration    uint64 `json:"duration"`
}

func (hb *HypeBot) setThemesong(file_path string, guild_id string, user_id string) string {
	if filePath, ok := hb.userStore.GetThemesong(guild_id, user_id); ok {
		// Delete old themesong
		if err := os.Remove(filePath); err != nil {
			log.Printf("failed to delete old themesong %s: %v", filePath, err)
		}
		return themesong.Update(hb.db, file_path, guild_id, user_id)
	}

	return themesong.Set(hb.db, file_path, guild_id, user_id)
}

func (hb *HypeBot) removeThemesong(guild_id string, user_id string) string {
	return themesong.Remove(hb.db, guild_id, user_id)
}

// runPlayer joins channelID and plays pb's queue until it drains, then leaves.
// The caller must have claimed pb with startPlayback. The guild stays claimed
// until the bot has left the channel, so a new join can never start a second
// connection mid-teardown; songs queued while leaving are played after
// rejoining.
func (hb *HypeBot) runPlayer(pb *playback, guildID, channelID string) {
	for {
		// The connection inherits the session's log level; writing it here
		// would race with the voice goroutines started by the join.
		ctx, cancel := context.WithTimeout(context.Background(), voiceTimeout)
		vc, err := hb.s.ChannelVoiceJoin(ctx, guildID, channelID, false, false)
		cancel()
		if err != nil {
			log.Printf("failed to join voice channel %s: %v", channelID, err)
			if vc != nil {
				leaveVoice(vc)
			}
			// Free the guild so the next join can try again.
			pb.reset()
			return
		}

		hb.playQueue(pb, vc)
		leaveVoice(vc)
		if !pb.finish() {
			return
		}
	}
}

// playQueue plays pb's queue on vc until it drains. A track that fails to
// play is skipped so it does not cost the songs queued behind it.
func (hb *HypeBot) playQueue(pb *playback, vc *discordgo.VoiceConnection) {
	for {
		track, ok := pb.current()
		if !ok {
			return
		}

		if err := hb.playTrack(vc, track); err != nil {
			log.Printf("skipping track %s: %v", track, err)
		} else {
			// Let the audio finish before the next track or leaving. Songs
			// queued meanwhile are still played.
			time.Sleep(POST_PLAYBACK_DELAY)
		}

		if !pb.advance() {
			return
		}
	}
}

func (hb *HypeBot) downloadVideo(url, start_time, duration string) (file_path string, err error) {
	// Acquire semaphore slot (blocks if 10 downloads already in progress)
	hb.downloadSem <- struct{}{}
	defer func() { <-hb.downloadSem }()

	var filePath string

	valid, err := hb.validateUrl(url)
	if !valid {
		log.Printf("It seems the url is not valid: %v", err)
		return "", err
	}

	ytdl, err := exec.LookPath("yt-dlp")
	if err != nil {
		log.Printf("It seems like yt-dlp was not found: %v", err)
		return "", err
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("It seems like the directory was not found: %v", err)
		return "", err
	}

	// Create songs directory if it doesn't exist
	songsDir := dir + "/songs/"
	if _, err := os.Stat(songsDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(songsDir, os.ModePerm)
		if err != nil {
			log.Printf("It seems like the directory could not be created: %v", err)
			return "", err
		}
	}
	fileName := uuid.New().String()
	fileNameComp := uuid.New().String()
	opusFile := songsDir + fileName + ".opus"
	opusFileComp := songsDir + fileNameComp + ".opus"
	filePath = fmt.Sprintf("./songs/%s.dca", fileName)

	args := buildArgs(url, fileName+".opus", start_time, duration)
	cmd := exec.Command(ytdl, args...)
	data, err := cmd.Output()
	if err != nil && !isYtdlpMaxDownloads(err) {
		log.Printf("{yt-dlp}-unhandled: %v (%v, %v, %v) \n", err, url, start_time, duration)
		return "", fmt.Errorf("There was an error processing your request ⚠️")
	}

	if len(data) < 1 {
		log.Printf("{yt-dlp}-no_data: %v (%v, %v, %v) \n", err, url, start_time, duration)
		return "", fmt.Errorf("Unable to retrieve requested audio ⚠️")
	}

	videoMetaData := VideoMetaData{}
	err = json.Unmarshal(data, &videoMetaData)
	if err != nil {
		return "", err
	}

	// Check for valid start time
	st, err := strconv.ParseUint(start_time, 10, 64)
	if err != nil {
		return "", err
	}

	if st > videoMetaData.Duration {
		return "", fmt.Errorf("Invalid start time ⚠️")
	}

	fmpg, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Printf("ffmpeg not found in $PATH: %v", err)
		return "", err
	}
	args = []string{
		"-i",
		opusFile,
		"-filter:a",
		"loudnorm",
		opusFileComp,
	}

	cmd = exec.Command(fmpg, args...)
	btt, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error while running ffmpeg: %v (%v)", err, string(btt))
		return "", err
	}

	// Convert opus to dca so we can send to discord voice
	log.Println("Converting " + fileName + ".opus to " + fileName + ".dca")
	encodeSession, _ := dca.EncodeFile(opusFileComp, dca.StdEncodeOptions)
	defer encodeSession.Cleanup()

	dcaFile, err := os.Create(songsDir + fileName + ".dca")
	if err != nil {
		log.Printf("failed to create dca file: %v", err)
		return "", err
	}

	_, err = io.Copy(dcaFile, encodeSession)
	dcaFile.Close() // Close before cleanup regardless of error
	if err != nil {
		log.Printf("unable to write to %s: %v", songsDir+fileName+".dca", err)
		return "", err
	}

	if err := os.Remove(opusFile); err != nil {
		log.Printf("failed to remove temp file %s: %v", opusFile, err)
	}

	if err := os.Remove(opusFileComp); err != nil {
		log.Printf("failed to remove temp file %s: %v", opusFileComp, err)
	}

	log.Printf("Created theme song: %v • %v.dca \n", videoMetaData.Title, fileName)

	return filePath, nil
}

func (hb *HypeBot) validateUrl(url string) (bool, error) {
	ytdl, err := exec.LookPath("yt-dlp")
	if err != nil {
		log.Printf("{yt-dlp}-not_found: %v", err)
		return false, fmt.Errorf("There was an error processing your command :warning:")
	}

	args := []string{
		url,
		"--extract-audio",
		"--ignore-errors",
		"--no-playlist",
		"--no-check-formats",
		"--print", "%(is_live)s|%(duration)s",
		"--simulate",
	}

	args = appendCommonYtdlpArgs(args)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(ytdl, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil && !isYtdlpMaxDownloads(err) {
		dlerr := parseYtdlpError(stderr.String())
		log.Printf("{yt-dlp}-validation_error: %v - %v\n", err, dlerr)
		return false, fmt.Errorf("%s :warning:", dlerr)
	}

	if err := checkVideoLimits(stdout.String()); err != nil {
		return false, err
	}

	return true, nil
}

// maxVideoSeconds is the longest video a themesong can be clipped from.
const maxVideoSeconds = 600

// checkVideoLimits rejects live videos and videos of maxVideoSeconds or more,
// given yt-dlp's "%(is_live)s|%(duration)s" output. yt-dlp prints NA for
// missing fields.
func checkVideoLimits(output string) error {
	line, _, _ := strings.Cut(strings.TrimSpace(output), "\n")
	isLive, duration, _ := strings.Cut(strings.TrimSpace(line), "|")

	if isLive == "True" {
		return fmt.Errorf("Unable to process a live video :warning:")
	}

	seconds, err := strconv.ParseFloat(duration, 64)
	if err != nil {
		log.Printf("{yt-dlp}-unknown_duration: %q", output)
		return fmt.Errorf("Unable to determine the video's length :warning:")
	}
	if seconds >= maxVideoSeconds {
		return fmt.Errorf("Video must not exceed 10 minutes :warning:")
	}

	return nil
}

// isYtdlpMaxDownloads reports whether yt-dlp exited with code 101, which it
// uses when --max-downloads stops it after a successful download.
func isYtdlpMaxDownloads(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 101
}

func parseYtdlpError(stderr string) string {
	for line := range strings.SplitSeq(stderr, "\n") {
		if strings.HasPrefix(line, "ERROR:") {
			parts := strings.SplitN(line, ": ", 3)
			if len(parts) >= 3 {
				return parts[2]
			}
			return strings.TrimPrefix(line, "ERROR: ")
		}
	}
	return "Unable to process video"
}

// appendCommonYtdlpArgs adds proxy, cookies, and POToken args used by both
// buildArgs and validateUrl.
func appendCommonYtdlpArgs(args []string) []string {
	if len(strings.TrimSpace(ProxyURL)) > 0 {
		args = append(args, "--proxy", ProxyURL)
	}

	args = append(args, "--cookies", "cookies.txt")

	if !DisablePOToken && len(strings.TrimSpace(POToken)) > 0 {
		args = append(args,
			"--extractor-args", fmt.Sprintf("youtube:player-client=web;po_token=web+%s", POToken),
		)
	}

	return args
}

func buildArgs(url, opusFile, start_time, duration string) []string {
	args := []string{
		url,
		"-v",
		"--extract-audio",
		"--ignore-errors",
		"--audio-format", "opus",
		"--max-downloads", "1",
		"--paths", "songs/",
		"--no-playlist",
		"--no-color",
		"--no-check-formats",
		"--print-json",
		"--quiet",
		"--output", opusFile,
		"--downloader", "ffmpeg",
		"--downloader-args", fmt.Sprintf("ffmpeg:-ss %s -t %s -b:a 96k", start_time, duration),
	}

	return appendCommonYtdlpArgs(args)
}

func (hb *HypeBot) playTrack(vc *discordgo.VoiceConnection, trackPath string) error {
	file, err := os.Open(trackPath)
	if err != nil {
		return fmt.Errorf("failed to open track %s: %w", trackPath, err)
	}
	defer file.Close()

	if err := vc.Speaking(true); err != nil {
		return fmt.Errorf("failed to set speaking state: %w", err)
	}
	defer vc.Speaking(false)

	return hb.streamAudio(vc, dca.NewDecoder(file))
}

func (hb *HypeBot) streamAudio(vc *discordgo.VoiceConnection, decoder *dca.Decoder) error {
	for {
		frame, err := decoder.OpusFrame()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("decode error: %w", err)
		}

		sent, err := sendFrame(vc, frame)
		if err != nil {
			return err
		}
		if !sent {
			return nil // Timed out, stop streaming
		}
	}
}

var errVoiceClosed = errors.New("voice connection closed")

// sendFrame sends one opus frame, reporting false if it timed out.
//
// discordgo closes OpusSend shortly after a connection dies (for example when
// the bot is disconnected from the channel), and a send on a closed channel
// panics. Checking Dead first avoids that in practice; the recover covers the
// remaining window, since the channel is closed by discordgo's goroutine.
func sendFrame(vc *discordgo.VoiceConnection, frame []byte) (sent bool, err error) {
	select {
	case <-vc.DeadChannel():
		return false, errVoiceClosed
	default:
	}

	defer func() {
		if recover() != nil {
			sent, err = false, errVoiceClosed
		}
	}()

	select {
	case vc.OpusSend <- frame:
		return true, nil
	case <-vc.DeadChannel():
		return false, errVoiceClosed
	case <-time.After(SEND_TIMEOUT):
		return false, nil
	}
}
