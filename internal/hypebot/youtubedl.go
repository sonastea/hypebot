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
	"sync"
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
		del := exec.Command("rm", filePath)
		del.Run()
		return themesong.Update(hb.db, file_path, guild_id, user_id)
	}

	return themesong.Set(hb.db, file_path, guild_id, user_id)
}

func (hb *HypeBot) removeThemesong(guild_id string, user_id string) string {
	return themesong.Remove(hb.db, guild_id, user_id)
}

func (hb *HypeBot) playThemesong(e *discordgo.VoiceStateUpdate, vc *discordgo.VoiceConnection) (err error) {
	guildID := e.VoiceState.GuildID
	channelID := e.ChannelID
	ctx := context.Background()

	for len(hb.getQueue(guildID, channelID)) > 0 {
		if err := hb.playNextTrack(vc, guildID, channelID); err != nil {
			return err
		}

		if !hb.advanceQueue(guildID, channelID) {
			vc.Disconnect(ctx)
			break
		}

		time.Sleep(POST_PLAYBACK_DELAY)
	}

	return nil
}

func (hb *HypeBot) downloadVideo(url, start_time, duration string) (file_path string, err error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
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
	if err != nil && err.Error() != "exit status 101" {
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
		log.Println(err)
	}

	_, err = io.Copy(dcaFile, encodeSession)
	if err != nil {
		log.Printf("unable to write to %s: %v", songsDir+fileName+".dca", err)
		return "", err
	}

	del := exec.Command("rm", opusFile)
	if del.Run() != nil {
		log.Println(err)
	}

	del = exec.Command("rm", opusFileComp)
	if del.Run() != nil {
		log.Println(err)
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
		"--match-filter", "!is_live & duration < 600",
		"--simulate",
	}

	if len(strings.TrimSpace(ProxyURL)) > 0 {
		args = append(args, "--proxy", ProxyURL)
	}

	args = append(args, "--cookies", "cookies.txt")

	// Add POToken if enabled
	if !DisablePOToken && len(strings.TrimSpace(POToken)) > 0 {
		args = append(args,
			"--extractor-args", fmt.Sprintf("youtube:player-client=web;po_token=web+%s", POToken),
		)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(ytdl, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil && err.Error() != "exit status 101" {
		dlerr := parseYtdlpError(stderr.String())
		log.Printf("{yt-dlp}-validation_error: %v - %v\n", err, dlerr)
		return false, fmt.Errorf("%s :warning:", dlerr)
	}

	output := stdout.String()
	if strings.Contains(output, "!is_live") {
		return false, fmt.Errorf("Unable to process a live video :warning:")
	}
	if strings.Contains(output, "duration < 600") {
		return false, fmt.Errorf("Video must not exceed 10 minutes :warning:")
	}

	return true, nil
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
		"--output", fmt.Sprintf("%s", opusFile),
		"--downloader", "ffmpeg",
		"--downloader-args", fmt.Sprintf("ffmpeg:-ss %s -t %s -b:a 96k", start_time, duration),
	}

	if len(strings.TrimSpace(ProxyURL)) > 0 {
		args = append(args, "--proxy", ProxyURL)
	}

	args = append(args, "--cookies", "cookies.txt")

	// Add POToken if enabled
	if !DisablePOToken && len(strings.TrimSpace(POToken)) > 0 {
		args = append(args,
			"--extractor-args", fmt.Sprintf("youtube:player-client=web;po_token=web+%s", POToken),
		)
	}

	return args
}

func (hb *HypeBot) getQueue(guildID, channelID string) []string {
	return hb.guildCacheStore[guildID].VCS[channelID]
}

func (hb *HypeBot) playNextTrack(vc *discordgo.VoiceConnection, guildID, channelID string) error {
	trackPath := hb.getQueue(guildID, channelID)[0]

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

		select {
		case vc.OpusSend <- frame:
		case <-time.After(SEND_TIMEOUT):
			return nil // Timed out, stop streaming
		}
	}
}

func (hb *HypeBot) advanceQueue(guildID, channelID string) bool {
	queue := hb.getQueue(guildID, channelID)

	if len(queue) > 1 {
		hb.guildCacheStore[guildID].VCS[channelID] = queue[1:]
		return true
	}

	// Last track finished
	time.Sleep(POST_PLAYBACK_DELAY)
	hb.guildCacheStore[guildID].VCS[channelID] = nil
	hb.guildCacheStore[guildID].Playing = false

	return false
}
