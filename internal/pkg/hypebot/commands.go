package hypebot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
	"github.com/sonastea/hypebot/internal/utils"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "clear",
			Description: "Clear your currently set themesong.",
		},
	}

	registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
)

func (hb *HypeBot) clearCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	exists, err := users.FindUser(hb.db, i.Member.User.ID)
	utils.CheckErr(err)

	if !exists {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have not set a themesong with HypeBot.",
			},
		})
	} else {
		msg = hb.removeThemesong(i.Member.User.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
	}
}
