package hypebot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/datastore/user"
)

func (hb *HypeBot) listenVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.VoiceState.UserID == BotID {
		return
	}

	// User enters a voice channel
	if e.BeforeUpdate == nil {
		fmt.Printf("%s:%s joined channel %s \n", e.Member.User.Username, e.VoiceState.UserID, e.ChannelID)
		// If user doesn't exist, add them to the database
		exists := hb.userStore.Find(e.GuildID, e.VoiceState.UserID)
		if !exists {
			newUser := &user.User{
				Guild_ID: e.GuildID,
				UID:      e.VoiceState.UserID,
			}
			hb.userStore.Add(*newUser)
		}

		vs, _ := s.State.VoiceState(e.GuildID, BotID)
		if hb.guildCacheStore[e.GuildID].Playing && vs.ChannelID != e.ChannelID {
			return
		}

		if filePath, ok := hb.userStore.GetThemesong(e.GuildID, e.VoiceState.UserID); ok {
			var vc *discordgo.VoiceConnection
			var err error

			hb.guildCacheStore[e.GuildID].VCS[e.ChannelID] = append(hb.guildCacheStore[e.GuildID].VCS[e.ChannelID], filePath)
			if !hb.guildCacheStore[e.GuildID].Playing {
				vc, err = hb.s.ChannelVoiceJoin(e.VoiceState.GuildID, e.ChannelID, false, false)
				vc.LogLevel = discordgo.LogInformational
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

	hb.guildStore.Add(e.ID)
	hb.guildCacheStore[e.ID] = hb.guildStore.Get(e.ID)

	log.Printf("Joined server `%s`:%s \n", e.Guild.Name, e.ID)
}

func (hb *HypeBot) listenOnLeaveServer(s *discordgo.Session, e *discordgo.GuildDelete) {
	g, found := hb.guildCacheStore[e.ID]
	if found {
		hb.guildStore.Remove(g.UID)
		delete(hb.guildCacheStore, g.UID)
		log.Printf("%s removed HypeBot. \n", g.UID)
	}
}
