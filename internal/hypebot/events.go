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
			hb.userStore.Add(user.User{
				Guild_ID: e.GuildID,
				UID:      e.VoiceState.UserID,
			})
		}

		filePath, ok := hb.userStore.GetThemesong(e.GuildID, e.VoiceState.UserID)
		if !ok {
			return
		}

		pb := hb.guilds.playback(e.GuildID)
		if pb.enqueue(e.ChannelID, filePath) != startPlayback {
			// Either the current player will play it, or the bot is busy in
			// another channel and the song is dropped.
			return
		}

		hb.runPlayer(pb, e.GuildID, e.ChannelID)
	}
}

func (hb *HypeBot) listenOnJoinServer(s *discordgo.Session, e *discordgo.GuildCreate) {
	// GuildCreate is also sent for every guild on (re)connect.
	if !hb.guilds.add(e.ID) {
		return
	}

	hb.guildStore.Add(e.ID)
	log.Printf("Joined server `%s`:%s \n", e.Guild.Name, e.ID)
}

func (hb *HypeBot) listenOnLeaveServer(s *discordgo.Session, e *discordgo.GuildDelete) {
	if hb.guilds.remove(e.ID) {
		hb.guildStore.Remove(e.ID)
		log.Printf("%s removed HypeBot. \n", e.ID)
	}
}
