package hypebot

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/datastore"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
)

// Variables used for command line parameters
var (
	Token           string
	POToken         string
	ProxyURL        string
	BotID           string
	GuildID         string
	DisableCommands string
	RemoveCommands  bool
)

type HypeBot struct {
	disabledCommands map[string]bool

	s  *discordgo.Session
	db *sql.DB

	guildCacheStore guild.CacheStore

	guildStore *guild.Store
	userStore  *user.Store
}

type (
	InteractionCreateHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Middleware               func(s *discordgo.Session, i *discordgo.InteractionCreate, next InteractionCreateHandler)
)

func setupEnv() {
	POToken = os.Getenv("POToken")
	ProxyURL = os.Getenv("PROXY_URL")

	if len(strings.TrimSpace(POToken)) > 0 {
		if _, err := os.Stat("cookies.txt"); os.IsNotExist(err) {
			panic("cookies.txt is required when using POToken")
		}
	}
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&BotID, "bid", "994803132259381291", "User ID of bot")
	flag.StringVar(&GuildID, "g", "", "Guild in which bot is running")
	flag.StringVar(&DisableCommands, "discmds", "", "Comma-separated list of commands to disable")
	flag.BoolVar(&RemoveCommands, "rmcmd", true, "Remove all commands after shutdowning or not")
}

func applyMiddlewares(handler InteractionCreateHandler, middlewares []Middleware) InteractionCreateHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := handler
		handler = InteractionCreateHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			mw(s, i, next)
		})
	}

	return handler
}

func NewHypeBot(db *sql.DB) (hb *HypeBot, err error) {
	flag.Parse()
	setupEnv()

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
		disabledCommands: make(map[string]bool, len(commands)),
		s:                dg,
		db:               db,
		guildCacheStore:  guild.NewGuildCacheStore(),
		guildStore:       guild.NewGuildStore(store),
		userStore:        user.NewUserStore(store),
	}, nil
}

func (hb *HypeBot) respondCmdProcessing(s *discordgo.Session, i *discordgo.InteractionCreate, name string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: fmt.Sprintf("Processing %s command...", name),
		},
	})
}

func (hb *HypeBot) respondCmdDisabled(s *discordgo.Session, i *discordgo.InteractionCreate, name string) error {
	message := fmt.Sprintf("The `%s` command is currently disabled 🚫.", name)
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong ❌",
		})
	}

	return err
}

func (hb *HypeBot) isCommandDisabled(name string) bool {
	return hb.disabledCommands[name]
}

func (hb *HypeBot) cmdCheckMiddleware(name string) Middleware {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate, next InteractionCreateHandler) {
		hb.respondCmdProcessing(s, i, name)

		if hb.isCommandDisabled(name) {
			hb.respondCmdDisabled(s, i, name)
			return
		}

		next(s, i)
	}
}

func (hb *HypeBot) disableCommands() {
	disabledCommands := make(map[string]bool, len(commands))
	if DisableCommands != "" {
		for _, cmd := range strings.Split(DisableCommands, ",") {
			disabledCommands[strings.TrimSpace(cmd)] = true
		}
	}

	hb.disabledCommands = disabledCommands
}

func (hb *HypeBot) handleCommands() {
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"clear": hb.clearCommand,
		"set":   hb.setCommand,
	}

	hb.disableCommands()

	hb.s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		name := i.ApplicationCommandData().Name

		if handler, ok := commandHandlers[name]; ok {
			readyHandler := applyMiddlewares(handler, []Middleware{
				hb.cmdCheckMiddleware(name),
			})

			readyHandler(s, i)
		}
	})

	for i, v := range commands {
		if hb.disabledCommands[v.Name] {
			log.Printf("Command `%s` is disabled.", v.Name)
		}

		cmd, err := hb.s.ApplicationCommandCreate(hb.s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func (hb *HypeBot) initGuildStore() {
	for _, g := range hb.s.State.Guilds {
		hb.guildStore.Add(g.ID)
		hb.guildCacheStore[g.ID] = hb.guildStore.Get(g.ID)
	}
}

func (hb *HypeBot) setCustomStatus() {
	err := hb.s.UpdateCustomStatus(os.Getenv("CUSTOM_STATUS"))
	if err != nil {
		log.Println("Error setting custom status:", err)
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
	hb.setCustomStatus()

	hb.s.AddHandler(hb.listenVoiceStateUpdate)
	hb.s.AddHandler(hb.listenOnJoinServer)
	hb.s.AddHandler(hb.listenOnLeaveServer)

	log.Printf("%v #%v is now running. Press CTRL-C to exit.\n", hb.s.State.User.Username, hb.s.State.User.Discriminator)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	return stop
}

func (hb *HypeBot) Stop() error {
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
				return err
			}
		}
	}

	log.Printf("%v #%v has gracefully shut down. \n", hb.s.State.User.Username, hb.s.State.User.Discriminator)

	return nil
}
