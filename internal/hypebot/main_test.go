package hypebot

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/hypebot/models"
	"github.com/stretchr/testify/assert"
)

var (
	hb  *HypeBot
	db  *sql.DB
	mhb *MockedHypeBot

	testBotToken = os.Getenv("TEST_TOKEN")
)

type MockedHypeBot struct {
	db         *sql.DB
	guilds     []*discordgo.Guild
	guildStore guild.Store
}

func TestMain(m *testing.M) {
	db, _ = sql.Open("sqlite3", ":memory:")

	_, err := db.Exec(database.Schema)
	if err != nil {
		log.Println(err)
	}

	mhb = &MockedHypeBot{
		db: db,
		guilds: []*discordgo.Guild{
			{
				ID: "mock",
			},
		},
		guildStore: make(map[string]*models.Guild),
	}

	os.Exit(m.Run())
}

func (mhb *MockedHypeBot) InitGuildStore() {
	for _, g := range mhb.guilds {
		guild.AddGuild(mhb.db, g.ID)
		mhb.guildStore[g.ID] = guild.GetGuild(mhb.db, g.ID)
	}
}

func TestInitGuildStore(t *testing.T) {
	mhb.InitGuildStore()
	assert.Exactly(t, 1, len(mhb.guildStore), "guild store should have 1 guild")

	for i, g := range mhb.guildStore {
		assert.Exactly(t, g.UID, mhb.guildStore[i].UID, "guild %v from guild store does not match %v", mhb.guildStore[i].UID, g.UID)
	}
}

func TestNewHypeBot(t *testing.T) {
	if !assert.NotEmpty(t, testBotToken) {
		t.Fatal("env(testBotToken), TEST_TOKEN not set")
	}

	if !assert.NotNil(t, db) {
		t.Fatal("unable to open an SQLITE memory database")
	}

	dg, err := discordgo.New("Bot " + testBotToken)
	if err != nil {
		t.Fatalf("unable to create a discord session: %s", err)
	}

	hb = &HypeBot{
		s:          dg,
		db:         db,
		guildStore: guild.NewGuildStore(),
	}

	assert.IsType(t, &HypeBot{}, hb)
}

func TestRunAndStop(t *testing.T) {
	botChan := hb.Run()
	assert.NotNil(t, botChan, "unable to return chan os.Signal from running hypebot")

	err := hb.Stop(botChan)
	assert.Nil(t, err, "unable to shut down hypebot gracefully: %v", err)
}
