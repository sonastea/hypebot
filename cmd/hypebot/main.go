package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/utils"
)

// Variables used for command line parameters
var (
	Token   string
	GuildID string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&GuildID, "g", "", "Guild in which bot is running")
	flag.Parse()
}

func main() {
	// Create discordgo session using a bot token
	dg, err := discordgo.New("Bot " + Token)
	utils.CheckErr(err)

	err = dg.Open()
	if err != nil {
		log.Println("Error opening connection:", err)
		return
	}
	dg.StateEnabled = true

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds)

	dg.AddHandler(listenToVoiceStateUpdate)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session after recieving CTRL-C signal
	defer dg.Close()
}

func listenToVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	// User enters a voice channel
	if e.BeforeUpdate == nil {
		fmt.Printf("%+v joined channel %+v \n\n", e.VoiceState.UserID, e.ChannelID)
	}
}
