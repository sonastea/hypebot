package hypebot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/hypebot/models"
)

func (hb *HypeBot) listenVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.VoiceState.UserID == BotID {
		return
	}

	// User enters a voice channel
	if e.BeforeUpdate == nil {
		fmt.Printf("%s:%s joined channel %s \n", e.Member.User.Username, e.VoiceState.UserID, e.ChannelID)
		// If user doesn't exist, add them to the database
		exists := hb.userStore.FindUser(hb.db, e.GuildID, e.VoiceState.UserID)
		if !exists {
			newUser := &models.User{
				Guild_ID: e.GuildID,
				UID:      e.VoiceState.UserID,
			}
			hb.userStore.AddUser(hb.db, *newUser)
		}

		vs, _ := s.State.VoiceState(e.GuildID, BotID)
		if hb.guildCacheStore[e.GuildID].Playing && vs.ChannelID != e.ChannelID {
			return
		}

		if filePath, ok := hb.userStore.GetThemesong(hb.db, e.GuildID, e.VoiceState.UserID); ok {
			var vc *discordgo.VoiceConnection
			var err error

			hb.guildCacheStore[e.GuildID].VCS[e.ChannelID] = append(hb.guildCacheStore[e.GuildID].VCS[e.ChannelID], filePath)
			if !hb.guildCacheStore[e.GuildID].Playing {
				vc, err = hb.s.ChannelVoiceJoin(e.VoiceState.GuildID, e.ChannelID, false, false)
				if err != nil {
					log.Println(err)
				}
			}

			if vc == nil || len(hb.guildCacheStore[e.GuildID].VCS[e.ChannelID]) > 1 {
				return
			}

			hb.guildCacheStore[e.GuildID].Playing = true
			err = hb.playThemesong(e, vc)
			if err != nil {
				log.Println(err)
				vc.Disconnect()
			}
		}
	}
}

func (hb *HypeBot) listenOnJoinServer(s *discordgo.Session, e *discordgo.GuildCreate) {
	_, found := hb.guildCacheStore[e.ID]
	if found {
		return
	}

	hb.guildStore.AddGuild(hb.db, e.ID)
	hb.guildCacheStore[e.ID] = hb.guildStore.GetGuild(hb.db, e.ID)

	log.Printf("Joined server `%s`:%s \n", e.Guild.Name, e.ID)
}

func (hb *HypeBot) listenOnLeaveServer(s *discordgo.Session, e *discordgo.GuildDelete) {
	g, found := hb.guildCacheStore[e.ID]
	if found {
		hb.guildStore.RemoveGuild(hb.db, g.UID)
		delete(hb.guildCacheStore, g.UID)
		log.Printf("%s removed HypeBot. \n", g.UID)
	}
}
