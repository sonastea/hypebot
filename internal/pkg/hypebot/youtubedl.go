package hypebot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jonas747/dca"
	"github.com/sonastea/hypebot/internal/pkg/datastore/themesongs"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
	"github.com/sonastea/hypebot/internal/utils"
)

type VideoMetaData struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	OriginalURL string `json:"original_url"`
	UploadDate  string `json:"upload_date"`
}

func (hb *HypeBot) setThemesong(file_path string, user_id string) string {
	if filePath, ok := users.GetThemesong(hb.db, user_id); ok {
		// Delete old themesong
		del := exec.Command("rm", filePath)
		del.Run()
		return themesongs.UpdateThemesong(hb.db, file_path, user_id)
	}

	return themesongs.SetThemesong(hb.db, file_path, user_id)
}

func (hb *HypeBot) removeThemesong(user_id string) string {
	return themesongs.RemoveThemesong(hb.db, user_id)
}

func (hb *HypeBot) playThemesong(file_path string, guild_id string, channel_id string, vc *discordgo.VoiceConnection) (err error) {
	if vc == nil {
		vc, err = hb.s.ChannelVoiceJoin(guild_id, channel_id, false, true)
	}
	if err != nil {
		return err
	}

	file, err := os.Open(file_path)
	utils.CheckErr(err)
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
	utils.CheckErr(err)

	err = vc.Disconnect()
	utils.CheckErr(err)

	return nil
}

func (hb *HypeBot) downloadVideo(url string, start_time string, duration string) (file_path string, err error) {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	var filePath string

	ytdl, err := exec.LookPath("yt-dlp")
	if err != nil {
		utils.CheckErrFatal(err)
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
		opusFile := songsDir + fileName + ".opus"
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
			log.Printf("{yt-dlp} %v\n", err)
		} else {
			videoMetaData := VideoMetaData{}
			err = json.Unmarshal(data, &videoMetaData)
			if err != nil {
				return "", err
			}

			// Convert opus to dca so we can send to discord voice
			fmt.Println("Converting " + fileName + ".mp3 to " + fileName + ".dca")
			encodeSession, _ := dca.EncodeFile(opusFile, dca.StdEncodeOptions)
			defer encodeSession.Cleanup()

			dcaFile, err := os.Create(songsDir + fileName + ".dca")
			utils.CheckErr(err)
			io.Copy(dcaFile, encodeSession)

			del := exec.Command("rm", opusFile)
			if del.Run() != nil {
				utils.CheckErr(err)
			}

			fmt.Printf("Created theme song: %v - %v \n", videoMetaData.Title, fileName)
		}
	}

	return filePath, nil
}
