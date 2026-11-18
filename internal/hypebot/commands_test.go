package hypebot

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type apiCall struct{ method, path, body string }

// fakeDiscordAPI routes the session's REST calls to a recorder. Edits of the
// original interaction response fail with editStatus.
func fakeDiscordAPI(t *testing.T, s *discordgo.Session, editStatus int) *[]apiCall {
	t.Helper()
	var calls []apiCall
	s.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
		}
		calls = append(calls, apiCall{r.Method, r.URL.Path, string(body)})
		status := http.StatusOK
		if r.Method == http.MethodPatch {
			status = editStatus
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Request:    r,
		}, nil
	})}
	return &calls
}

func setInteraction(url, start, duration string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Type:    discordgo.InteractionApplicationCommand,
		AppID:   "app",
		Token:   "token",
		GuildID: "guild-a",
		Member:  &discordgo.Member{User: &discordgo.User{ID: "user-a", Username: "user-a"}},
		Data: discordgo.ApplicationCommandInteractionData{
			Name:    "set",
			Options: makeCommandOptions(url, start, duration),
		},
	}}
}

func TestSetCommandReportsInvalidInput(t *testing.T) {
	for _, tt := range []struct {
		name       string
		editStatus int
		wantCalls  []string
	}{
		{"edits the reply", http.StatusOK, []string{http.MethodPatch}},
		{"falls back to a follow-up when the edit fails", http.StatusBadRequest, []string{http.MethodPatch, http.MethodPost}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, err := discordgo.New("Bot test")
			require.NoError(t, err)
			s.MaxRestRetries = 0
			calls := fakeDiscordAPI(t, s, tt.editStatus)

			(&HypeBot{}).setCommand(s, setInteraction("https://example.com/video", "0s", "5"))

			var methods []string
			for _, c := range *calls {
				methods = append(methods, c.method)
				assert.Contains(t, c.path, "/webhooks/app/token")
			}
			require.Equal(t, tt.wantCalls, methods)
			assert.Contains(t, (*calls)[0].body, "is not a valid url")
			if len(*calls) > 1 {
				var followup struct {
					Content string `json:"content"`
					Flags   int    `json:"flags"`
				}
				require.NoError(t, json.Unmarshal([]byte((*calls)[1].body), &followup))
				assert.Equal(t, "Something went wrong", followup.Content)
				assert.Equal(t, int(discordgo.MessageFlagsEphemeral), followup.Flags&int(discordgo.MessageFlagsEphemeral),
					"the fallback is only visible to the user who ran the command")
			}
		})
	}
}

func TestSanitizeSetCommand(t *testing.T) {
	const watchURL = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	for _, tt := range []struct {
		name, url, start, duration, wantSeconds string
	}{
		{"watch URL with minutes and seconds", watchURL, "9m45s", "10", "585"},
		{"short URL with zero start", "https://youtu.be/dQw4w9WgXcQ", "0s", "5", "0"},
		{"seconds-only start and maximum duration", watchURL, "45s", "15", "45"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			url, start, duration, err := sanitizeSetCommand(makeCommandOptions(tt.url, tt.start, tt.duration))
			require.NoError(t, err)
			assert.Equal(t, tt.url, url)
			assert.Equal(t, tt.wantSeconds, start)
			assert.Equal(t, tt.duration, duration)
		})
	}

	for _, tt := range []struct {
		name, url, start, duration, wantError string
	}{
		{"invalid URL", "https://example.com/video", "0s", "5", "is not a valid url"},
		{"invalid start time", watchURL, "invalid", "5", "invalid duration"},
		{"invalid duration", watchURL, "0s", "abc", "is not a valid time"},
		{"zero duration", watchURL, "0s", "0", "can't be less than"},
		{"negative duration", watchURL, "0s", "-5", "can't be less than"},
		{"over maximum duration", watchURL, "0s", "16", "greater than"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := sanitizeSetCommand(makeCommandOptions(tt.url, tt.start, tt.duration))
			require.ErrorContains(t, err, tt.wantError)
		})
	}
}
