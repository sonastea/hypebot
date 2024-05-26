package hypebot

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/datastore"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
)

// Variables used for command line parameters
var (
	Token          string
	BotID          string
	GuildID        string
	RemoveCommands bool
)

type HypeBot struct {
	s  *discordgo.Session
	db *sql.DB

	guildCacheStore guild.CacheStore

	guildStore *guild.Store
	userStore  *user.Store
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&BotID, "bid", "994803132259381291", "User ID of bot")
	flag.StringVar(&GuildID, "g", "", "Guild in which bot is running")
	flag.BoolVar(&RemoveCommands, "rmcmd", true, "Remove all commands after shutdowning or not")
}

func NewHypeBot(db *sql.DB) (hb *HypeBot, err error) {
	flag.Parse()

	// Create discordgo session using a bot token
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		return nil, err
	}

	store, err := datastore.New(db)
	if err != nil {
		return nil, err
	}

	return &HypeBot{
		s:               dg,
		db:              db,
		guildCacheStore: guild.NewGuildCacheStore(),
		guildStore:      guild.NewGuildStore(store),
		userStore:       user.NewUserStore(store),
	}, nil
}

func (hb *HypeBot) initGuildStore() {
	for _, g := range hb.s.State.Guilds {
		hb.guildStore.Add(g.ID)
		hb.guildCacheStore[g.ID] = hb.guildStore.Get(g.ID)
	}
}

func (hb *HypeBot) handleCommands() {
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"clear": hb.clearCommand,
		"set":   hb.setCommand,
	}

	hb.s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	for i, v := range commands {
		cmd, err := hb.s.ApplicationCommandCreate(hb.s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func (hb *HypeBot) Run() chan os.Signal {
	// Create websocket connection to discord with the discord session
	err := hb.s.Open()
	if err != nil {
		log.Println("Error opening connection:", err)
		return nil
	}
	hb.s.StateEnabled = true
	hb.s.LogLevel = discordgo.LogInformational

	hb.s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds)

	hb.initGuildStore()
	hb.handleCommands()

	hb.s.AddHandler(hb.listenVoiceStateUpdate)
	hb.s.AddHandler(hb.listenOnJoinServer)
	hb.s.AddHandler(hb.listenOnLeaveServer)

	log.Printf("HypeBot #%v is now running. Press CTRL-C to exit.\n", hb.s.State.User.Discriminator)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	return stop
}

func (hb *HypeBot) Stop(sig chan os.Signal) error {
	close(sig)

	var err error

	err = hb.s.Close()
	if err != nil {
		return err
	}

	err = hb.db.Close()
	if err != nil {
		return err
	}

	if RemoveCommands {
		for _, v := range registeredCommands {
			err := hb.s.ApplicationCommandDelete(hb.s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Printf("HypeBot #%v has gracefully shut down. \n", hb.s.State.User.Discriminator)

	return nil
}
