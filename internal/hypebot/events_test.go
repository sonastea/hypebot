package hypebot

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/themesong"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testGuild    = "guild-a"
	botChannel   = "channel-bot"
	otherChannel = "channel-other"
)

// newVoiceTestBot returns a bot whose guild is already playing "current.dca"
// in botChannel. Its session has no gateway, so voice joins fail immediately.
func newVoiceTestBot(t *testing.T) *HypeBot {
	t.Helper()
	s, err := discordgo.New("Bot test")
	require.NoError(t, err)
	s.VoiceConnections = make(map[string]*discordgo.VoiceConnection) // normally set by Open

	store := testutils.NewStore(t)
	hb := &HypeBot{
		s:          s,
		db:         store.DB(),
		guilds:     newGuilds(),
		guildStore: guild.NewGuildStore(store),
		userStore:  user.NewUserStore(store),
	}
	require.Equal(t, startPlayback, hb.guilds.playback(testGuild).enqueue(botChannel, "current.dca"))
	return hb
}

func voiceEvent(guildID, userID, channelID string, before *discordgo.VoiceState) *discordgo.VoiceStateUpdate {
	return &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			GuildID:   guildID,
			UserID:    userID,
			ChannelID: channelID,
			Member:    &discordgo.Member{User: &discordgo.User{ID: userID, Username: userID}},
		},
		BeforeUpdate: before,
	}
}

func TestVoiceJoinRegistersMembers(t *testing.T) {
	hb := newVoiceTestBot(t)

	hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, BotID, botChannel, nil))
	assert.False(t, hb.userStore.Find(testGuild, BotID), "the bot's own voice updates are ignored")

	hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, "user-a", otherChannel, &discordgo.VoiceState{ChannelID: botChannel}))
	assert.False(t, hb.userStore.Find(testGuild, "user-a"), "moving between channels is not a join")

	hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, "user-a", botChannel, nil))
	hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, "user-a", botChannel, nil))
	assert.True(t, hb.userStore.Find(testGuild, "user-a"))
	total, ok := hb.userStore.GetTotalServed()
	require.True(t, ok)
	assert.Equal(t, uint64(1), total, "rejoining does not register the member twice")

	_, _, queue := hb.guilds.playback(testGuild).state()
	assert.Equal(t, []string{"current.dca"}, queue, "members without a themesong queue nothing")
	assert.Empty(t, hb.s.VoiceConnections)
}

func TestVoiceJoinQueuesThemesongs(t *testing.T) {
	hb := newVoiceTestBot(t)
	require.Equal(t, themesong.SONG_ADDED_SUCCESS, themesong.Set(hb.db, "user-a.dca", testGuild, "user-a"))
	require.Equal(t, themesong.SONG_ADDED_SUCCESS, themesong.Set(hb.db, "user-b.dca", testGuild, "user-b"))

	hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, "user-a", botChannel, nil))
	hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, "user-b", otherChannel, nil))

	playing, channel, queue := hb.guilds.playback(testGuild).state()
	assert.True(t, playing)
	assert.Equal(t, botChannel, channel)
	assert.Equal(t, []string{"current.dca", "user-a.dca"}, queue,
		"a join in the bot's channel waits behind the current track; a join elsewhere is dropped")
	assert.Empty(t, hb.s.VoiceConnections, "a busy bot does not open a second voice connection")
}

func TestVoiceJoinFailureLeavesGuildIdle(t *testing.T) {
	hb := newVoiceTestBot(t)
	hb.guilds.playback(testGuild).reset()
	require.Equal(t, themesong.SONG_ADDED_SUCCESS, themesong.Set(hb.db, "user-a.dca", testGuild, "user-a"))

	for range 2 {
		hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, "user-a", otherChannel, nil))

		playing, _, queue := hb.guilds.playback(testGuild).state()
		assert.False(t, playing, "a failed join must not leave the guild marked as playing")
		assert.Empty(t, queue, "a failed join must not leave its song queued")
		assert.Empty(t, hb.s.VoiceConnections, "a connection that could not leave cleanly is dropped")
	}
}

func TestGuildReaddKeepsPlayback(t *testing.T) {
	hb := newVoiceTestBot(t)
	hb.listenOnJoinServer(hb.s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: testGuild}})
	pb := hb.guilds.playback(testGuild)

	// The bot is removed and re-added while its player still owns the guild.
	hb.listenOnLeaveServer(hb.s, &discordgo.GuildDelete{Guild: &discordgo.Guild{ID: testGuild}})
	hb.listenOnJoinServer(hb.s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: testGuild}})

	assert.Same(t, pb, hb.guilds.playback(testGuild))
	assert.Equal(t, busy, hb.guilds.playback(testGuild).enqueue(otherChannel, "other.dca"),
		"no second player can start while the first still owns the guild")
}

func TestVoiceJoinInUnknownGuild(t *testing.T) {
	hb := newVoiceTestBot(t)
	require.Equal(t, themesong.SONG_ADDED_SUCCESS, themesong.Set(hb.db, "user-a.dca", "guild-new", "user-a"))

	// The voice event arrives before the bot has registered the guild.
	assert.NotPanics(t, func() {
		hb.listenVoiceStateUpdate(hb.s, voiceEvent("guild-new", "user-a", botChannel, nil))
	})
	assert.True(t, hb.userStore.Find("guild-new", "user-a"))
	playing, _, _ := hb.guilds.playback("guild-new").state()
	assert.False(t, playing, "the failed join leaves the new guild idle")

	// Joining the guild is still processed and persisted afterwards.
	assert.False(t, hb.guilds.has("guild-new"), "a voice event is not a guild membership")
	hb.listenOnJoinServer(hb.s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: "guild-new"}})
	assert.True(t, hb.guilds.has("guild-new"))
	g := hb.guildStore.Get("guild-new")
	require.NotNil(t, g)
	assert.Equal(t, "guild-new", g.UID, "the guild is saved to the database")
}

func TestConcurrentVoiceJoins(t *testing.T) {
	hb := newVoiceTestBot(t)
	const members = 20
	for i := range members {
		for _, ch := range []string{botChannel, otherChannel} {
			id := fmt.Sprintf("%s-user-%d", ch, i)
			require.Equal(t, themesong.SONG_ADDED_SUCCESS, themesong.Set(hb.db, id+".dca", testGuild, id))
		}
	}

	// Discord delivers events concurrently, while the bot also joins and
	// leaves other guilds.
	var wg sync.WaitGroup
	for i := range members {
		for _, ch := range []string{botChannel, otherChannel} {
			wg.Go(func() {
				hb.listenVoiceStateUpdate(hb.s, voiceEvent(testGuild, fmt.Sprintf("%s-user-%d", ch, i), ch, nil))
			})
		}
		other := fmt.Sprintf("guild-%d", i)
		wg.Go(func() {
			hb.listenOnJoinServer(hb.s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: other}})
			hb.listenOnLeaveServer(hb.s, &discordgo.GuildDelete{Guild: &discordgo.Guild{ID: other}})
		})
	}
	wg.Wait()

	playing, channel, queue := hb.guilds.playback(testGuild).state()
	assert.True(t, playing)
	assert.Equal(t, botChannel, channel)
	require.Len(t, queue, members+1, "every join in the bot's channel is queued exactly once")
	assert.Equal(t, "current.dca", queue[0], "the playing track stays first")
	want := []string{"current.dca"}
	for i := range members {
		want = append(want, fmt.Sprintf("%s-user-%d.dca", botChannel, i))
	}
	assert.ElementsMatch(t, want, queue)
	assert.Empty(t, hb.s.VoiceConnections, "no second player starts")
}

func TestServerMembership(t *testing.T) {
	hb := newVoiceTestBot(t)

	hb.listenOnJoinServer(hb.s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: "guild-b"}})
	hb.listenOnJoinServer(hb.s, &discordgo.GuildCreate{Guild: &discordgo.Guild{ID: testGuild}})
	assert.True(t, hb.guilds.has("guild-b"))
	require.NotNil(t, hb.guildStore.Get("guild-b"))
	assert.Equal(t, int8(1), hb.guildStore.Get("guild-b").Active)
	playing, _, _ := hb.guilds.playback(testGuild).state()
	assert.True(t, playing, "a repeated GuildCreate keeps the guild's playback")

	hb.listenOnLeaveServer(hb.s, &discordgo.GuildDelete{Guild: &discordgo.Guild{ID: "guild-b"}})
	assert.False(t, hb.guilds.has("guild-b"))
	assert.Equal(t, int8(0), hb.guildStore.Get("guild-b").Active, "leaving marks the guild inactive")
}
