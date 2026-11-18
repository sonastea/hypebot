package hypebot

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withTestCookies(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	require.NoError(t, os.WriteFile("cookies.txt", nil, 0o600))
}

func restoreEnvConfig(t *testing.T) {
	t.Helper()
	token, proxy, disabled := POToken, ProxyURL, DisableCommands
	t.Cleanup(func() { POToken, ProxyURL, DisableCommands = token, proxy, disabled })
}

func TestSetupEnv(t *testing.T) {
	withTestCookies(t)
	restoreEnvConfig(t)
	t.Setenv("POToken", "test-token")
	t.Setenv("PROXY_URL", "http://proxy.example")
	t.Setenv("DISABLED_COMMANDS", " set, , clear ")
	setupEnv()

	assert.Equal(t, "test-token", POToken)
	assert.Equal(t, "http://proxy.example", ProxyURL)

	hb := &HypeBot{}
	hb.disableCommands()
	assert.True(t, hb.isCommandDisabled("set"))
	assert.True(t, hb.isCommandDisabled("clear"))
	assert.False(t, hb.isCommandDisabled(""))
	assert.False(t, hb.isCommandDisabled("other"))
}

func TestApplyMiddlewaresRunsInOrder(t *testing.T) {
	var calls []string
	record := func(name string) Middleware {
		return func(s *discordgo.Session, i *discordgo.InteractionCreate, next InteractionCreateHandler) {
			calls = append(calls, name+":before")
			next(s, i)
			calls = append(calls, name+":after")
		}
	}
	stop := func(s *discordgo.Session, i *discordgo.InteractionCreate, next InteractionCreateHandler) {
		calls = append(calls, "stop")
	}
	handler := func(s *discordgo.Session, i *discordgo.InteractionCreate) { calls = append(calls, "handler") }

	applyMiddlewares(handler, []Middleware{record("outer"), record("inner")})(nil, nil)
	assert.Equal(t, []string{"outer:before", "inner:before", "handler", "inner:after", "outer:after"}, calls)

	calls = nil
	applyMiddlewares(handler, []Middleware{record("outer"), stop, record("inner")})(nil, nil)
	assert.Equal(t, []string{"outer:before", "stop", "outer:after"}, calls, "a middleware can short-circuit the chain")

	calls = nil
	applyMiddlewares(handler, nil)(nil, nil)
	assert.Equal(t, []string{"handler"}, calls)
}

func TestDisablePOTokenFlag(t *testing.T) {
	f := flag.Lookup("disable_potoken")
	require.NotNil(t, f)
	assert.Equal(t, "true", f.DefValue)

	previous := f.Value.String()
	t.Cleanup(func() { require.NoError(t, f.Value.Set(previous)) })
	require.NoError(t, f.Value.Set("true"))
	assert.True(t, DisablePOToken)
	require.NoError(t, f.Value.Set("false"))
	assert.False(t, DisablePOToken, "the CLI flag should enable use of the token when false")
}

func TestInitGuildStore(t *testing.T) {
	state := discordgo.NewState()
	for _, id := range []string{"guild-a", "guild-b"} {
		require.NoError(t, state.GuildAdd(&discordgo.Guild{ID: id}))
	}
	hb := &HypeBot{
		s:          &discordgo.Session{State: state},
		guilds:     newGuilds(),
		guildStore: guild.NewGuildStore(testutils.NewStore(t)),
	}

	// The gateway keeps updating state while startup reads it.
	started, stop, gatewayDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(gatewayDone)
		close(started)
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				state.GuildAdd(&discordgo.Guild{ID: fmt.Sprintf("guild-late-%d", i)})
			}
		}
	}()
	<-started
	hb.initGuildStore()
	close(stop)
	<-gatewayDone

	for _, id := range []string{"guild-a", "guild-b"} {
		assert.True(t, hb.guilds.has(id), "startup should register guild %s", id)
		persisted := hb.guildStore.Get(id)
		require.NotNil(t, persisted)
		assert.Equal(t, id, persisted.UID, "startup should persist guild %s", id)
	}
}

func TestRunAndStop(t *testing.T) {
	withTestCookies(t)
	restoreEnvConfig(t)
	t.Setenv("DISABLED_COMMANDS", "")

	token := Token
	t.Cleanup(func() { Token = token })
	Token = os.Getenv("TEST_TOKEN")
	require.NotEmpty(t, Token, "TEST_TOKEN is required for the Discord integration test")

	db := testutils.NewStore(t).DB()
	hb, err := NewHypeBot(db)
	require.NoError(t, err)
	botChan := hb.Run()
	require.NotNil(t, botChan, "HypeBot should connect and provide a shutdown signal channel")
	t.Cleanup(func() {
		signal.Stop(botChan)
		assert.NoError(t, hb.Stop())
		assert.Error(t, db.Ping(), "shutdown should close the bot's database")
	})

	registered, err := hb.s.ApplicationCommands(hb.s.State.User.ID, GuildID)
	require.NoError(t, err)
	names := make([]string, 0, len(registered))
	for _, cmd := range registered {
		names = append(names, cmd.Name)
	}
	assert.ElementsMatch(t, []string{"clear", "set"}, names)
}
