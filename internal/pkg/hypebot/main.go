package hypebot

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

type HypeBot struct {
	s *discordgo.Session
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&GuildID, "g", "", "Guild in which bot is running")
	flag.Parse()
}

func NewHypeBot() (hb *HypeBot, err error) {
	// Create discordgo session using a bot token
	dg, err := discordgo.New("Bot " + Token)
	utils.CheckErr(err)

	return &HypeBot{dg}, nil
}

func (hb *HypeBot) Run() {
	err := hb.s.Open()
	if err != nil {
		log.Println("Error opening connection:", err)
		return
	}
	hb.s.StateEnabled = true

	hb.s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds)

	hb.s.AddHandler(listenToVoiceStateUpdate)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session after recieving CTRL-C signal
	defer hb.s.Close()
}

func listenToVoiceStateUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	// User enters a voice channel
	if e.BeforeUpdate == nil {
		fmt.Printf("%+v joined channel %+v \n\n", e.VoiceState.UserID, e.ChannelID)
	}
}
