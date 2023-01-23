package hypebot

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/sonastea/hypebot/internal/pkg/datastore"
	"github.com/sonastea/hypebot/internal/utils"
)

// Variables used for command line parameters
var (
	Token          string
	GuildID        string
	RemoveCommands bool
)

type HypeBot struct {
	s  *discordgo.Session
	db *sql.DB
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&GuildID, "g", "", "Guild in which bot is running")
	flag.BoolVar(&RemoveCommands, "rmcmd", true, "Remove all commands after shutdowning or not")
	flag.Parse()
}

func NewHypeBot() (hb *HypeBot, err error) {
	// Create discordgo session using a bot token
	dg, err := discordgo.New("Bot " + Token)
	utils.CheckErrFatal(err)

	db, err := datastore.NewDBConn()
	utils.CheckErrFatal(err)

	return &HypeBot{
		s:  dg,
		db: db,
	}, nil
}

func (hb *HypeBot) handleCommands() {
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"clear": hb.clearCommand,
	}

	hb.s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	for i, v := range commands {
		// log.Println(i, v, GuildID, RemoveCommands)
		cmd, err := hb.s.ApplicationCommandCreate(hb.s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func (hb *HypeBot) Run() {
	// Cleanly close down the Discord session after recieving CTRL-C signal
	defer func() {
		hb.cleanup()
	}()

	// Create websocket connection to discord with the discord session
	err := hb.s.Open()
	if err != nil {
		log.Println("Error opening connection:", err)
		return
	}
	hb.s.StateEnabled = true

	hb.s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds)

	hb.s.AddHandler(hb.listenVoiceStateUpdate)
	hb.s.AddHandler(hb.listenMessageCreate)

	hb.handleCommands()

	fmt.Printf("HypeBot #%v is now running. Press CTRL-C to exit.\n", hb.s.State.User.Discriminator)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-stop
}

func (hb *HypeBot) cleanup() {
	hb.s.Close()
	hb.db.Close()

	if RemoveCommands {
		for _, v := range registeredCommands {
			err := hb.s.ApplicationCommandDelete(hb.s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}
