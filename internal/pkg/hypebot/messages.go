package hypebot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	msg                  string
	available_messages   string = "Available HypeBot commands are ```hypebot <set | remove>```"
	not_enough_arguments string = "Invalid command, use ```hypebot set <youtube_url | video_id> start_time (duration: min: 1s, max: 15s)```"
)

func (hb *HypeBot) handleMessage(s *discordgo.Session, m *discordgo.Message) {
	// Clean message contents of excess whitespace before splitting
	str := strings.Join(strings.Fields(m.Content), " ")
	args := strings.Split(str, " ")

	// check for hypebot prefix for commands
	if args[0] != "hypebot" {
		return
	}

	// check if supplied arguments after hypebot prefix are valid
	if len(args) < 2 {
		s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
		return
	}

	// run available commands accorrding to the supplied arguments
	switch args[1] {
	case "set":
		if len(args) < 4 {
			s.ChannelMessageSendReply(m.ChannelID, not_enough_arguments, m.Reference())
			return
		}

		url, start, duration, err := sanitizeSetMessage(args[2:])
		if err != nil {
			s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
		}

		filePath, err := hb.downloadVideo(*url, start, duration)
		if err != nil {
			s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
		}

		msg = hb.setThemesong(filePath, m.Author.ID)

	case "remove":
		msg = hb.removeThemesong(m.Author.ID)

	default:
		msg = "Invalid command or arguments"
	}

	s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
}

func sanitizeSetMessage(args []string) (*string, string, string, error) {
	validUrl, _ := regexp.MatchString(`(?:https?:\/\/)?(?:www\.)?youtu\.?be(?:\.com)?\/?.*(?:watch|embed)?(?:.*v=|v\/|\/)([\w\-_]+)\&?`, args[0])
	if !validUrl {
		return nil, "", "", fmt.Errorf("Invalid youtube url")
	}

	_, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return nil, "", "", fmt.Errorf("Invalid start time, use a number in seconds")
	}

	if len(args) < 3 {
		return &args[0], args[1], "5", nil
	}

	return &args[0], args[1], args[2], nil
}
