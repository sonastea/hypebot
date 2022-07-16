package hypebot

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jonas747/dca"
	"github.com/kkdai/youtube/v2"
	"github.com/sonastea/hypebot/internal/pkg/datastore/themesongs"
	"github.com/sonastea/hypebot/internal/utils"
)

type YoutubeDL struct {
	mu sync.Mutex
	c  *youtube.Client
}

func NewYoutubeDL() *YoutubeDL {
	return &YoutubeDL{
		c: &youtube.Client{},
	}
}

func (hb *HypeBot) handleSetThemesong(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	var (
		videoID   string
		startTime float64
		duration  float64 = 5.0
	)

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)

	if opt, ok := optionMap["url"]; ok {
		url := opt.StringValue()
		validURL := strings.Contains(url, "youtube.com/watch?v=")
		if !validURL {
			return "Please provide a valid youtube url from the search bar"
		}
		videoID = strings.Split(url, "=")[1]
	} else {
		return "Please provide a valid youtube url from the search bar"
	}

	if opt, ok := optionMap["start_time"]; ok {
		startTime = opt.FloatValue()
	} else {
		return "Please provide a starting time in seconds"
	}

	if opt, ok := optionMap["duration"]; ok {
		duration = opt.FloatValue()
	}

	filename, err := hb.downloadVideo(videoID, startTime, duration)
	utils.CheckErr(err)
	filename += ".dca"

	go themesongs.SetThemesong(hb.db, i.User.ID, filename)

	return "Themesong set successfully"
}

func (hb *HypeBot) playThemesong(guildID string, channelID string, vc *discordgo.VoiceConnection) (err error) {
	if vc == nil {
		vc, err = hb.s.ChannelVoiceJoin(guildID, channelID, false, true)
	}
	if err != nil {
		return err
	}

	file, err := os.Open("./songs/4b1191c2-aca0-4e2e-b49e-cf5a010454f8.dca")
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

func (hb *HypeBot) downloadVideo(videoID string, startTime float64, duration float64) (string, error) {
	y := NewYoutubeDL()
	y.mu.Lock()
	defer y.mu.Unlock()

	// Get video from youtube
	video, err := y.c.GetVideo(videoID)
	utils.CheckErr(err)

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := y.c.GetStream(video, &formats[0])
	utils.CheckErr(err)

	dir, err := os.Getwd()
	utils.CheckErr(err)

	// Create songs directory if it doesn't exist
	songsDir := dir + "/songs/"
	if _, err := os.Stat(songsDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(songsDir, os.ModePerm)
		utils.CheckErr(err)
	}

	// Files we convert from mp4 -> mp3 -> dca
	fileName := uuid.New().String()
	mp3File := songsDir + fileName + ".mp3"
	videoFile, err := os.Create(songsDir + fileName + ".mp4")
	utils.CheckErr(err)
	defer videoFile.Close()

	// Copy content from youtube stream to mp4 file
	_, err = io.Copy(videoFile, stream)
	utils.CheckErr(err)

	// Convert mp4 to mp3
	fmt.Println("Converting " + fileName + ".mp4 to " + fileName + ".mp3")
	con := exec.Command("ffmpeg", "-ss", "60", "-t", "15", "-i", videoFile.Name(), mp3File)
	if con.Run() != nil {
		utils.CheckErr(err)
	}

	// Convert mp3 to dca so we can send to discord voice
	fmt.Println("Converting " + fileName + ".mp3 to " + fileName + ".dca")
	encodeSession, _ := dca.EncodeFile(mp3File, dca.StdEncodeOptions)
	defer encodeSession.Cleanup()
	dcaFile, err := os.Create(songsDir + fileName + ".dca")
	utils.CheckErr(err)
	io.Copy(dcaFile, encodeSession)

	// Delete mp3 and mp3 files after we're done
	del := exec.Command("rm", videoFile.Name())
	if del.Run() != nil {
		utils.CheckErr(err)
	}

	del2 := exec.Command("rm", mp3File)
	if del2.Run() != nil {
		utils.CheckErr(err)
	}

	fmt.Printf("Created theme song: %v - %v \n\n", video.Title, fileName)
	return fileName, nil
}
