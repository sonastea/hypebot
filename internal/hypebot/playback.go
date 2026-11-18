package hypebot

import (
	"context"
	"errors"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// voiceTimeout bounds how long joining or leaving a voice channel may take.
const voiceTimeout = 10 * time.Second

// enqueueResult tells the caller of playback.enqueue what to do next.
type enqueueResult int

const (
	// startPlayback means the caller claimed the guild and must join the
	// channel and play the queue.
	startPlayback enqueueResult = iota
	// queued means the song will be played by the guild's current player.
	queued
	// busy means the guild is playing in another channel; the song is dropped.
	busy
)

// playback is one guild's themesong queue. A guild plays in one voice channel
// at a time; the goroutine that receives startPlayback from enqueue is the
// only player until it calls finish (after leaving the channel) or reset.
//
// All methods are safe for concurrent use. They never block on I/O, so the
// lock is only held for in-memory bookkeeping.
type playback struct {
	mu        sync.Mutex
	playing   bool
	channelID string
	queue     []string
}

// enqueue requests that song be played in channelID.
func (p *playback) enqueue(channelID, song string) enqueueResult {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.playing {
		if p.channelID != channelID {
			return busy
		}
		p.queue = append(p.queue, song)
		return queued
	}

	p.playing = true
	p.channelID = channelID
	p.queue = []string{song}
	return startPlayback
}

// current returns the song that should be playing now.
func (p *playback) current() (song string, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.queue) == 0 {
		return "", false
	}
	return p.queue[0], true
}

// advance drops the song that just finished and reports whether more songs
// remain. The guild stays claimed either way; the player must leave the
// channel and then call finish.
func (p *playback) advance() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.queue) > 0 {
		p.queue = p.queue[1:]
	}
	return len(p.queue) > 0
}

// finish is called by the player after it has left the voice channel. If
// songs were queued while it was leaving, the guild stays claimed and finish
// reports true so the player rejoins and plays them; otherwise the guild
// becomes idle.
func (p *playback) finish() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.queue) > 0 {
		return true
	}
	p.resetLocked()
	return false
}

// reset empties the queue and marks the guild idle, dropping pending songs.
// The player uses it when it cannot join the channel at all.
func (p *playback) reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.resetLocked()
}

func (p *playback) resetLocked() {
	p.playing = false
	p.channelID = ""
	p.queue = nil
}

// state returns a snapshot of the playback state.
func (p *playback) state() (playing bool, channelID string, queue []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing, p.channelID, slices.Clone(p.queue)
}

// guilds tracks the guilds the bot is a member of and each guild's playback.
// Membership and playback are tracked separately: a voice event can need a
// guild's playback before the bot has processed joining that guild.
type guilds struct {
	mu      sync.Mutex
	members map[string]bool
	players map[string]*playback
}

func newGuilds() *guilds {
	return &guilds{
		members: make(map[string]bool),
		players: make(map[string]*playback),
	}
}

// add records membership of a guild, reporting whether it is new.
func (g *guilds) add(guildID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.members[guildID] {
		return false
	}
	g.members[guildID] = true
	return true
}

// remove forgets a guild's membership, reporting whether the bot was a member.
// The guild's playback is kept: a player may still own it, and replacing it
// would let a second player start in the same guild if the bot is re-added.
func (g *guilds) remove(guildID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.members[guildID] {
		return false
	}
	delete(g.members, guildID)
	return true
}

// has reports whether the bot is a member of a guild.
func (g *guilds) has(guildID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.members[guildID]
}

// playback returns the guild's playback, creating it on first use.
func (g *guilds) playback(guildID string) *playback {
	g.mu.Lock()
	defer g.mu.Unlock()

	p, ok := g.players[guildID]
	if !ok {
		p = &playback{}
		g.players[guildID] = p
	}
	return p
}

// leaveAckTimeout bounds how long leaveVoice keeps waiting for Discord to
// acknowledge a leave after voiceTimeout has passed.
const leaveAckTimeout = time.Minute

// leaveVoice disconnects from a voice channel. The caller must keep the guild
// claimed until it returns: discordgo applies a leave acknowledgment to
// whichever connection the guild has when it arrives, so an acknowledgment
// arriving after a new join would kill the new connection.
//
// If the acknowledgment is slow, leaveVoice keeps waiting up to
// leaveAckTimeout. If it never arrives, or the leave could not be sent, the
// connection is killed locally so discordgo forgets it.
func leaveVoice(vc *discordgo.VoiceConnection) {
	ctx, cancel := context.WithTimeout(context.Background(), voiceTimeout)
	err := vc.Disconnect(ctx)
	cancel()
	if err == nil {
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		select {
		case <-vc.DeadChannel(): // acknowledged late
			return
		case <-time.After(leaveAckTimeout):
		}
	}

	log.Printf("failed to leave voice channel cleanly, dropping the connection: %v", err)
	vc.Kill()
}
