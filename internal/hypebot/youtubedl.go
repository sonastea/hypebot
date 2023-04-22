package hypebot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/robrotheram/dca"
	"github.com/sonastea/hypebot/internal/datastore/themesong"
	"github.com/sonastea/hypebot/internal/datastore/user"
)

type VideoMetaData struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	OriginalURL string `json:"original_url"`
	UploadDate  string `json:"upload_date"`
	Duration    uint64 `json:"duration"`
}

func (hb *HypeBot) setThemesong(file_path string, guild_id string, user_id string) string {
	if filePath, ok := user.GetThemesong(hb.db, guild_id, user_id); ok {
		// Delete old themesong
		del := exec.Command("rm", filePath)
		del.Run()
		return themesong.UpdateThemesong(hb.db, file_path, guild_id, user_id)
	}

	return themesong.SetThemesong(hb.db, file_path, guild_id, user_id)
}

func (hb *HypeBot) removeThemesong(guild_id string, user_id string) string {
	return themesong.RemoveThemesong(hb.db, guild_id, user_id)
}

func (hb *HypeBot) playThemesong(e *discordgo.VoiceStateUpdate, vc *discordgo.VoiceConnection) (err error) {
	for len(hb.guildStore[e.VoiceState.GuildID].VCS[e.ChannelID]) > 0 {
		file, err := os.Open(hb.guildStore[e.VoiceState.GuildID].VCS[e.ChannelID][0])
		if err != nil {
			log.Println(err)
		}
		defer file.Close()

		decoder := dca.NewDecoder(file)
		_ = vc.Speaking(true)
		for {
			frame, err := decoder.OpusFrame()
			if err != nil {
				if err != io.EOF {
					return err
				}
				break
			}
			select {
			case vc.OpusSend <- frame:
			case <-time.After(time.Second * 5):
				break
			}
		}

		err = vc.Speaking(false)
		if err != nil {
			log.Println(err)
		}

		if len(hb.guildStore[e.GuildID].VCS[e.ChannelID]) > 1 {
			hb.guildStore[e.GuildID].VCS[e.ChannelID] = hb.guildStore[e.GuildID].VCS[e.ChannelID][1:]
		} else if len(hb.guildStore[e.GuildID].VCS[e.ChannelID]) == 1 {
			time.Sleep(1500 * time.Millisecond)
			hb.guildStore[e.GuildID].VCS[e.ChannelID] = hb.guildStore[e.VoiceState.GuildID].VCS[e.ChannelID][:0]
			hb.guildStore[e.GuildID].Playing = false
			vc.Disconnect()
		}

		time.Sleep(1500 * time.Millisecond)
	}

	return nil
}

func (hb *HypeBot) downloadVideo(url string, start_time string, duration string) (file_path string, err error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	var filePath string

	valid, err := hb.validateUrl(url)
	if !valid {
		return "", err
	}

	ytdl, err := exec.LookPath("yt-dlp")
	if err != nil {
		return "", err
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}

		// Create songs directory if it doesn't exist
		songsDir := dir + "/songs/"
		if _, err := os.Stat(songsDir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(songsDir, os.ModePerm)
			return "", err
		}
		fileName := uuid.New().String()
		fileNameComp := uuid.New().String()
		opusFile := songsDir + fileName + ".opus"
		opusFileComp := songsDir + fileNameComp + ".opus"
		filePath = fmt.Sprintf("./songs/%s.dca", fileName)

		args := []string{
			url,
			"--extract-audio",
			"--ignore-errors",
			"--audio-format", "opus",
			"--max-downloads", "1",
			"--no-playlist",
			"--no-color",
			"--no-check-formats",
			"--print-json",
			"--quiet",
			"--output", fmt.Sprintf("%s", opusFile),
			"--downloader", "ffmpeg",
			"--downloader-args", fmt.Sprintf("ffmpeg:-ss %s -t %s -b:a 96k", start_time, duration),
		}

		cmd := exec.Command(ytdl, args...)
		if data, err := cmd.Output(); err != nil && err.Error() != "exit status 101" {
			log.Printf("{yt-dlp}-unhandled: %v (%v, %v, %v) \n", err, url, start_time, duration)
			return "", errors.New("There was an error processing your request ⚠️")
		} else {
			if len(data) < 1 {
				log.Printf("{yt-dlp}-no_data: %v (%v, %v, %v) \n", err, url, start_time, duration)
				return "", errors.New("Unable to retrieve requested audio ⚠️")
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

			if st < 0 || st > videoMetaData.Duration {
				return "", errors.New("Invalid start time ⚠️")
			}

			fmpg, err := exec.LookPath("ffmpeg")
			if err != nil {
				return "", err
			}
			args := []string{
				"-i",
				opusFile,
				"-filter:a",
				"loudnorm",
				opusFileComp,
			}

			cmd = exec.Command(fmpg, args...)
			btt, err := cmd.CombinedOutput()
			if err != nil {
				log.Println(string(btt))
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
			io.Copy(dcaFile, encodeSession)

			del := exec.Command("rm", opusFile)
			if del.Run() != nil {
				log.Println(err)
			}

			del = exec.Command("rm", opusFileComp)
			if del.Run() != nil {
				log.Println(err)
			}

			log.Printf("Created theme song: %v • %v.dca \n", videoMetaData.Title, fileName)
		}
	}

	return filePath, nil
}

func (hb *HypeBot) validateUrl(url string) (valid bool, err error) {
	ytdl, err := exec.LookPath("yt-dlp")
	if err != nil {
		return false, err
	}

	args := []string{
		url,
		"--extract-audio",
		"--ignore-errors",
		"--no-playlist",
		"--no-check-formats",
		"--match-filter",
		"!is_live",
		"--simulate",
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(ytdl, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil && err.Error() != "exit status 101" {
		dlerr := string(bytes.Split(stderr.Bytes(), []byte(":"))[2])
		log.Printf("{yt-dlp}-not_live: %v -%v \n", err, dlerr)
		return false, errors.New(fmt.Sprint(dlerr, ":warning:"))
	}

	if bytes.Contains(stdout.Bytes(), []byte("!is_live")) {
		return false, errors.New("Unable to process a live video :warning:")
	}

	args[6] = "duration < 600"
	cmd = exec.Command(ytdl, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil && err.Error() != "exit status 101" {
		dlerr := string(bytes.Split(stderr.Bytes(), []byte(":"))[2])
		log.Printf("{yt-dlp}-not_live: %v -%v \n", err, dlerr)
		return false, errors.New(fmt.Sprint(dlerr, ":warning:"))
	}

	if bytes.Contains(stdout.Bytes(), []byte("duration < 600")) {
		return false, errors.New("Video must not exceed 10 minutes :warning:")
	}

	return true, nil
}
