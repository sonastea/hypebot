package hypebot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var msg string

func handleMessage(s *discordgo.Session, m *discordgo.Message) {
	args := strings.Split(m.Content, " ")
	// check for hypebot prefix for commands
	if args[0] != "hypebot" {
		return
	}

	// check if supplied arguments after hypebot prefix are valid
	if len(args) < 2 {
		msg = "Available HypeBot commands are ```hypebot <set | remove>```"
		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}

	// run available commands accorrding to the supplied arguments
	switch args[1] {
	case "set":
		msg = "set"
	case "remove":
		msg = "removed"
	default:
		msg = "Invalid command or arguments. Check spacing"
	}

	s.ChannelMessageSend(m.ChannelID, msg)
}
