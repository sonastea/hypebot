package hypebot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
	"github.com/sonastea/hypebot/internal/utils"
)

func (hb *HypeBot) listenVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.VoiceState.UserID == "994803132259381291" {
		return
	}

	// User enters a voice channel
	if e.BeforeUpdate == nil {
		fmt.Printf("%+v joined channel %+v \n", e.VoiceState.UserID, e.ChannelID)

		// If user doesn't exist, add them to the database
		exists, err := users.FindUser(hb.db, e.VoiceState.UserID)
		utils.CheckErr(err)

		if !exists {
			newUser := users.User{
				UID: e.VoiceState.UserID,
			}
			users.AddUser(hb.db, newUser)
		}

		if filePath, ok := users.GetThemesong(hb.db, e.VoiceState.UserID); ok {
			vc, err := hb.s.ChannelVoiceJoin(GuildID, e.ChannelID, false, false)
			if err != nil {
				log.Println(err)
			}

			err = hb.playThemesong(filePath, GuildID, e.ChannelID, vc)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
