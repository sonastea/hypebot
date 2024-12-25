package hypebot

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var (
	hb        *HypeBot
	db        *sql.DB
	mhb       *MockedHypeBot
	mockStore *testutils.MockStore

	testBotToken = os.Getenv("TEST_TOKEN")
)

type MockedHypeBot struct {
	db     *sql.DB
	guilds []*discordgo.Guild

	guildCacheStore guild.CacheStore

	guildStore *guild.Store
	userStore  *user.Store
}

func TestMain(m *testing.M) {
	db, _ = sql.Open("sqlite3", ":memory:")

	_, err := db.Exec(database.Schema)
	if err != nil {
		log.Println(err)
	}

	mockStore = testutils.NewMockStore(db)

	mhb = &MockedHypeBot{
		db: db,
		guilds: []*discordgo.Guild{
			{
				ID: "mock",
			},
		},
		guildCacheStore: make(map[string]*guild.Guild),
		guildStore:      guild.NewGuildStore(mockStore),
		userStore:       user.NewUserStore(mockStore),
	}

	os.Exit(m.Run())
}

func (mhb *MockedHypeBot) InitGuildStore() {
	for _, g := range mhb.guilds {
		mhb.guildStore.Add(g.ID)
		mhb.guildCacheStore[g.ID] = mhb.guildStore.Get(g.ID)
	}
}

func TestSetupEnv(t *testing.T) {
	tempFile, err := os.Create("cookies.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary cookies.txt file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	t.Setenv("POToken", "POTokenValue")
	setupEnv()

	assert.Equalf(t, "POTokenValue", POToken, "POToken should be POTokenValue")
}

func TestInitGuildStore(t *testing.T) {
	mhb.InitGuildStore()
	assert.Exactly(t, 1, len(mhb.guildCacheStore), "guild store should have 1 guild")

	for i, g := range mhb.guildCacheStore {
		assert.Exactly(t, g.UID, mhb.guildCacheStore[i].UID, "guild %v from guild store does not match %v", mhb.guildCacheStore[i].UID, g.UID)
	}
}

func TestIsCommandDisabled(t *testing.T) {
	cmdName := "test-name"
	hb := &HypeBot{
		disabledCommands: map[string]bool{cmdName: true},
	}

	cmd := hb.isCommandDisabled(cmdName)
	assert.Exactly(t, true, cmd, "cmd should be true, but cmd is %v", cmd)
}

func TestDisableCommands(t *testing.T) {
	os.Args = []string{"./hypebot", "--discmds=clear,set"}
	flag.Parse()

	hb := &HypeBot{disabledCommands: make(map[string]bool, 2)}
	hb.disableCommands()

	expected := map[string]bool{"clear": true, "set": true}
	assert.Exactly(t, expected, hb.disabledCommands, "disabled commands should only have 'set'")

	DisableCommands = ""
}

func TestHandleCommands(t *testing.T) {
	dg, err := discordgo.New("Bot " + testBotToken)
	if err != nil {
		t.Fatalf("unable to create a discord session: %s", err)
	}

	hb = &HypeBot{
		s:               dg,
		db:              db,
		guildCacheStore: guild.NewGuildCacheStore(),
		guildStore:      guild.NewGuildStore(mockStore),
		userStore:       user.NewUserStore(mockStore),
	}

	err = hb.s.Open()
	if err != nil {
		t.Fatalf("unable opening discord websocket connection: %s", err)
	}

	hb.handleCommands()
	expectedCommands := len(registeredCommands)
	if cmds, err := hb.s.ApplicationCommands("1083967553887555595", ""); err == nil {
		assert.Equalf(t, expectedCommands, len(cmds), "application commands should have %s registered commands: got %s want %s", expectedCommands, len(cmds), expectedCommands)
	} else {
		t.Fatalf("unable to retrieve discord commands: %s", err)
	}

	err = hb.s.Close()
	if err != nil {
		t.Fatalf("unable closing discord websocket connection: %s", err)
	}
}

// func TestSetCustomStatus(t *testing.T) {
// 	dg, err := discordgo.New("Bot " + testBotToken)
// 	if err != nil {
// 		t.Fatalf("unable to create a discord session: %s", err)
// 	}
//
// 	hb = &HypeBot{
// 		s:               dg,
// 		db:              db,
// 		guildCacheStore: guild.NewGuildCacheStore(),
// 		guildStore:      guild.NewGuildStore(mockStore),
// 		userStore:       user.NewUserStore(mockStore),
// 	}
//
// 	err = hb.s.Open()
// 	if err != nil {
// 		t.Fatalf("unable opening discord websocket connection: %s", err)
// 	}
//
// 	expectedCustomStatus := "custom status"
// 	os.Setenv("CUSTOM_STATUS", expectedCustomStatus)
//
// 	hb.setCustomStatus()
// 	assert.Equal(t, expectedCustomStatus, hb.s.Identify.Presence.Status, "custom status should be %s, but is %s", expectedCustomStatus, hb.s.Identify.Presence.Status)
//
// 	err = hb.s.Close()
// 	if err != nil {
// 		t.Fatalf("unable closing discord websocket connection: %s", err)
// 	}
// }

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
		s:               dg,
		db:              db,
		guildCacheStore: guild.NewGuildCacheStore(),
		guildStore:      guild.NewGuildStore(mockStore),
		userStore:       user.NewUserStore(mockStore),
	}

	assert.IsType(t, &HypeBot{}, hb)
}

func TestRunAndStop(t *testing.T) {
	botChan := hb.Run()
	assert.NotNil(t, botChan, "unable to return chan os.Signal from running hypebot")

	err := hb.Stop()
	assert.Nil(t, err, "unable to shut down hypebot gracefully: %v", err)
}
