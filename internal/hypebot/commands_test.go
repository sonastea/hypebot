package hypebot

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func makeCommandOptions(url, startTime, duration string) []*discordgo.ApplicationCommandInteractionDataOption {
	return []*discordgo.ApplicationCommandInteractionDataOption{
		{
			Name:  "song-url",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: url,
		},
		{
			Name:  "start-time",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: startTime,
		},
		{
			Name:  "duration",
			Type:  discordgo.ApplicationCommandOptionString,
			Value: duration,
		},
	}
}

func TestSanitizeSetCommand(t *testing.T) {
	t.Run("valid YouTube URL with standard format", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "1m30s", "10")
		url, start, dur, err := sanitizeSetCommand(opts)

		assert.NoError(t, err)
		assert.Equal(t, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", url)
		assert.Equal(t, "90", start) // 1m30s = 90 seconds
		assert.Equal(t, "10", dur)
	})

	t.Run("valid YouTube URL with short format", func(t *testing.T) {
		opts := makeCommandOptions("https://youtu.be/dQw4w9WgXcQ", "0s", "5")
		url, start, dur, err := sanitizeSetCommand(opts)

		assert.NoError(t, err)
		assert.Equal(t, "https://youtu.be/dQw4w9WgXcQ", url)
		assert.Equal(t, "0", start)
		assert.Equal(t, "5", dur)
	})

	t.Run("valid start time in seconds only", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "45s", "3")
		_, start, _, err := sanitizeSetCommand(opts)

		assert.NoError(t, err)
		assert.Equal(t, "45", start)
	})

	t.Run("valid start time with minutes and seconds", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "9m45s", "5")
		_, start, _, err := sanitizeSetCommand(opts)

		assert.NoError(t, err)
		assert.Equal(t, "585", start) // 9*60 + 45 = 585
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		opts := makeCommandOptions("https://example.com/video", "0s", "5")
		_, _, _, err := sanitizeSetCommand(opts)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is not a valid url")
	})

	t.Run("invalid start time format returns error", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "invalid", "5")
		_, _, _, err := sanitizeSetCommand(opts)

		assert.Error(t, err)
	})

	t.Run("duration less than or equal to 0 returns error", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "0s", "0")
		_, _, _, err := sanitizeSetCommand(opts)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't be less than")
	})

	t.Run("negative duration returns error", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "0s", "-5")
		_, _, _, err := sanitizeSetCommand(opts)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can't be less than")
	})

	t.Run("duration greater than 15 returns error", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "0s", "16")
		_, _, _, err := sanitizeSetCommand(opts)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "greater than")
	})

	t.Run("duration at max boundary (15) succeeds", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "0s", "15")
		_, _, dur, err := sanitizeSetCommand(opts)

		assert.NoError(t, err)
		assert.Equal(t, "15", dur)
	})

	t.Run("invalid duration format returns error", func(t *testing.T) {
		opts := makeCommandOptions("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "0s", "abc")
		_, _, _, err := sanitizeSetCommand(opts)

		assert.Error(t, err)
	})
}
