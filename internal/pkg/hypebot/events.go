package hypebot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/pkg/datastore/guilds"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
)

func (hb *HypeBot) listenVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.VoiceState.UserID == "994803132259381291" || e.VoiceState.UserID == "1083967553887555595" {
		return
	}

	// User enters a voice channel
	if e.BeforeUpdate == nil {
		fmt.Printf("%+v joined channel %+v \n", e.VoiceState.UserID, e.ChannelID)

		// If user doesn't exist, add them to the database
		exists := users.FindUser(hb.db, e.GuildID, e.VoiceState.UserID)

		if !exists {
			newUser := &users.User{
				Guild_ID: e.GuildID,
				UID:      e.VoiceState.UserID,
			}
			users.AddUser(hb.db, *newUser)
		}

		if filePath, ok := users.GetThemesong(hb.db, e.GuildID, e.VoiceState.UserID); ok {
			vc, err := hb.s.ChannelVoiceJoin(e.VoiceState.GuildID, e.ChannelID, false, false)
			if err != nil {
				log.Println(err)
			}

			hb.guildStore[e.GuildID].VCS[e.ChannelID] = append(hb.guildStore[e.GuildID].VCS[e.ChannelID], filePath)

			err = hb.playThemesong(e, e.ChannelID, vc)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (hb *HypeBot) listenOnJoinServer(s *discordgo.Session, e *discordgo.GuildCreate) {
	_, found := hb.guildStore[e.ID]
	if found {
		return
	}

	guilds.AddGuild(hb.db, e.ID)
	hb.guildStore[e.ID] = guilds.GetGuild(hb.db, e.ID)

	log.Printf("Joined server `%v`:%v \n", e.Guild.Name, e.ID)
}

func (hb *HypeBot) listenOnLeaveServer(s *discordgo.Session, e *discordgo.GuildDelete) {
	guild, found := hb.guildStore[e.ID]
	if found {
		guilds.RemoveGuild(hb.db, guild.UID)
		delete(hb.guildStore, guild.UID)
		log.Printf("%v removed HypeBot. \n", guild.UID)
	}
}
