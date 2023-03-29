package hypebot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/pkg/datastore/guilds"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
	"github.com/sonastea/hypebot/internal/utils"
)

func (hb *HypeBot) listenVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.VoiceState.UserID == BotID {
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

		vs, err := s.State.VoiceState(e.GuildID, BotID)
		if err != nil {
			utils.CheckErr(nil)
		}

		if hb.guildStore[e.GuildID].Playing && vs.ChannelID != e.ChannelID {
			return
		}

		if filePath, ok := users.GetThemesong(hb.db, e.GuildID, e.VoiceState.UserID); ok {
			var vc *discordgo.VoiceConnection
			var err error

            hb.guildStore[e.GuildID].VCS[e.ChannelID] = append(hb.guildStore[e.GuildID].VCS[e.ChannelID], filePath)
			if !hb.guildStore[e.GuildID].Playing {
				vc, err = hb.s.ChannelVoiceJoin(e.VoiceState.GuildID, e.ChannelID, false, false)
				if err != nil {
					utils.CheckErr(err)
				}
			}

			if len(hb.guildStore[e.VoiceState.GuildID].VCS[e.ChannelID]) > 1 {
				return
			}

			hb.guildStore[e.GuildID].Playing = true
			err = hb.playThemesong(e, e.ChannelID, vc)
			if err != nil {
				utils.CheckErr(err)
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
