package hypebot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/robrotheram/dca"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dcaFrames encodes frames in the length-prefixed format read by dca.Decoder.
func dcaFrames(t *testing.T, frames ...[]byte) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	for _, f := range frames {
		require.NoError(t, binary.Write(&buf, binary.LittleEndian, int16(len(f))))
		buf.Write(f)
	}
	return &buf
}

func TestPlaybackQueue(t *testing.T) {
	var pb playback

	require.Equal(t, startPlayback, pb.enqueue(botChannel, "first.dca"), "an idle guild starts playing")
	require.Equal(t, queued, pb.enqueue(botChannel, "second.dca"))
	require.Equal(t, busy, pb.enqueue(otherChannel, "other.dca"), "one channel at a time")

	track, ok := pb.current()
	require.True(t, ok)
	assert.Equal(t, "first.dca", track)

	require.True(t, pb.advance(), "more tracks remain")
	track, _ = pb.current()
	assert.Equal(t, "second.dca", track)

	require.False(t, pb.advance(), "last track finished")
	_, ok = pb.current()
	assert.False(t, ok)

	// The player is now leaving the channel; the guild stays claimed.
	assert.Equal(t, queued, pb.enqueue(botChannel, "late.dca"), "a join while leaving waits for the player")
	assert.Equal(t, busy, pb.enqueue(otherChannel, "other.dca"), "no second player starts while leaving")
	require.True(t, pb.finish(), "the player rejoins for songs queued while leaving")
	playing, channel, queue := pb.state()
	assert.True(t, playing)
	assert.Equal(t, botChannel, channel)
	assert.Equal(t, []string{"late.dca"}, queue)

	require.False(t, pb.advance())
	require.False(t, pb.finish(), "nothing left: the guild becomes idle")
	playing, channel, queue = pb.state()
	assert.False(t, playing)
	assert.Empty(t, channel)
	assert.Empty(t, queue)

	assert.Equal(t, startPlayback, pb.enqueue(otherChannel, "other.dca"), "an idle guild can play in any channel")
	pb.reset()
	assert.Equal(t, startPlayback, pb.enqueue(botChannel, "again.dca"), "reset frees the guild")
}

func TestPlaybackConcurrentEnqueue(t *testing.T) {
	var pb playback
	const callers = 50
	results := make(chan enqueueResult, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Go(func() { results <- pb.enqueue(botChannel, fmt.Sprintf("song-%d.dca", i)) })
	}
	wg.Wait()
	close(results)

	starts := 0
	for r := range results {
		if r == startPlayback {
			starts++
		}
	}
	assert.Equal(t, 1, starts, "exactly one caller becomes the player")
	_, _, queue := pb.state()
	assert.Len(t, queue, callers)
}

func TestStreamAudio(t *testing.T) {
	hb := &HypeBot{}

	t.Run("sends every frame in order", func(t *testing.T) {
		vc := &discordgo.VoiceConnection{OpusSend: make(chan []byte, 3)}
		frames := dcaFrames(t, []byte("one"), []byte("two"), []byte("three"))

		require.NoError(t, hb.streamAudio(vc, dca.NewDecoder(frames)))
		close(vc.OpusSend)
		var got []string
		for f := range vc.OpusSend {
			got = append(got, string(f))
		}
		assert.Equal(t, []string{"one", "two", "three"}, got)
	})

	t.Run("stops at a truncated frame", func(t *testing.T) {
		vc := &discordgo.VoiceConnection{OpusSend: make(chan []byte, 2)}
		frames := dcaFrames(t, []byte("one"))
		require.NoError(t, binary.Write(frames, binary.LittleEndian, int16(10)))
		frames.WriteString("tw") // header promises 10 bytes, file ends after 2

		require.ErrorContains(t, hb.streamAudio(vc, dca.NewDecoder(frames)), "decode error")
		require.Len(t, vc.OpusSend, 1, "the partial frame is not sent")
		assert.Equal(t, []byte("one"), <-vc.OpusSend)
	})
}

func TestStreamAudioStopsOnClosedConnection(t *testing.T) {
	closed := func() chan []byte { c := make(chan []byte); close(c); return c }
	deadCh := make(chan struct{})
	close(deadCh)

	for _, tt := range []struct {
		name string
		vc   *discordgo.VoiceConnection
	}{
		// discordgo marks the connection dead, then closes OpusSend 100ms later.
		{"connection marked dead", &discordgo.VoiceConnection{OpusSend: make(chan []byte), Dead: deadCh}},
		{"send channel already closed", &discordgo.VoiceConnection{OpusSend: closed(), Dead: deadCh}},
		{"send channel closed before dead is observed", &discordgo.VoiceConnection{OpusSend: closed(), Dead: make(chan struct{})}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			require.NotPanics(t, func() {
				err = (&HypeBot{}).streamAudio(tt.vc, dca.NewDecoder(dcaFrames(t, []byte("one"), []byte("two"))))
			})
			assert.ErrorIs(t, err, errVoiceClosed)
		})
	}
}

func TestPlayQueueSkipsBadTracks(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.dca")
	unplayable := filepath.Join(dir, "unplayable.dca")
	require.NoError(t, os.WriteFile(unplayable, nil, 0o600))

	hb := newVoiceTestBot(t)
	var pb playback
	require.Equal(t, startPlayback, pb.enqueue(botChannel, missing))
	require.Equal(t, queued, pb.enqueue(botChannel, unplayable))

	// Without a gateway the join fails, but it still returns a voice
	// connection; playing on it fails, like a dropped connection would.
	vc, err := hb.s.ChannelVoiceJoin(t.Context(), testGuild, botChannel, false, false)
	require.Error(t, err)
	require.NotNil(t, vc)

	hb.playQueue(&pb, vc)

	playing, channel, queue := pb.state()
	assert.Empty(t, queue, "every queued track was attempted instead of stopping at the first failure")
	assert.True(t, playing, "the guild stays claimed until the player has left the channel")
	assert.Equal(t, botChannel, channel)
	assert.False(t, pb.finish(), "after leaving, the guild is freed")
}

func TestSetThemesongReplacesPreviousFile(t *testing.T) {
	store := testutils.NewStore(t)
	hb := &HypeBot{db: store.DB(), userStore: user.NewUserStore(store)}
	dir := t.TempDir()
	first, second := filepath.Join(dir, "first.dca"), filepath.Join(dir, "second.dca")
	require.NoError(t, os.WriteFile(first, nil, 0o600))
	require.NoError(t, os.WriteFile(second, nil, 0o600))

	assert.Equal(t, "Added themesong ✅", hb.setThemesong(first, testGuild, "user-a"))
	assert.Equal(t, "Updated themesong ✅", hb.setThemesong(second, testGuild, "user-a"))

	path, ok := hb.userStore.GetThemesong(testGuild, "user-a")
	require.True(t, ok)
	assert.Equal(t, second, path)
	assert.NoFileExists(t, first, "the replaced song is deleted from disk")
	assert.FileExists(t, second)

	assert.Equal(t, "Removed themesong ❌", hb.removeThemesong(testGuild, "user-a"))
	_, ok = hb.userStore.GetThemesong(testGuild, "user-a")
	assert.False(t, ok)
}
