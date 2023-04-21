package hypebot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/hypebot/models"
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
		exists := user.FindUser(hb.db, e.GuildID, e.VoiceState.UserID)
		if !exists {
			newUser := &models.User{
				Guild_ID: e.GuildID,
				UID:      e.VoiceState.UserID,
			}
			user.AddUser(hb.db, *newUser)
		}

		vs, err := s.State.VoiceState(e.GuildID, BotID)
		if err != nil {
			utils.CheckErr(nil)
		}

		if hb.guildStore[e.GuildID].Playing && vs.ChannelID != e.ChannelID {
			return
		}

		if filePath, ok := user.GetThemesong(hb.db, e.GuildID, e.VoiceState.UserID); ok {
			var vc *discordgo.VoiceConnection
			var err error

			hb.guildStore[e.GuildID].VCS[e.ChannelID] = append(hb.guildStore[e.GuildID].VCS[e.ChannelID], filePath)
			if !hb.guildStore[e.GuildID].Playing {
				vc, err = hb.s.ChannelVoiceJoin(e.VoiceState.GuildID, e.ChannelID, false, false)
				if err != nil {
					utils.CheckErr(err)
				}
			}

			if vc == nil || len(hb.guildStore[e.GuildID].VCS[e.ChannelID]) > 1 {
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

	guild.AddGuild(hb.db, e.ID)
	hb.guildStore[e.ID] = guild.GetGuild(hb.db, e.ID)

	log.Printf("Joined server `%v`:%v \n", e.Guild.Name, e.ID)
}

func (hb *HypeBot) listenOnLeaveServer(s *discordgo.Session, e *discordgo.GuildDelete) {
	g, found := hb.guildStore[e.ID]
	if found {
		guild.RemoveGuild(hb.db, g.UID)
		delete(hb.guildStore, g.UID)
		log.Printf("%v removed HypeBot. \n", g.UID)
	}
}
