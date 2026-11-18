package hypebot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYtdlpError(t *testing.T) {
	for _, tt := range []struct {
		name, stderr, want string
	}{
		{"extractor error after a warning", "WARNING: retrying\nERROR: [youtube] abc123: Video unavailable", "Video unavailable"},
		{"plain error", "ERROR: Something went wrong", "Something went wrong"},
		{"no error line", "[youtube] some other output", "Unable to process video"},
		{"no output", "", "Unable to process video"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseYtdlpError(tt.stderr))
		})
	}
}

func TestBuildArgs(t *testing.T) {
	token, proxy, disabled := POToken, ProxyURL, DisablePOToken
	t.Cleanup(func() { POToken, ProxyURL, DisablePOToken = token, proxy, disabled })

	const (
		videoURL  = "https://youtu.be/dQw4w9WgXcQ"
		proxyURL  = "http://proxy.example"
		testToken = "test-token"
	)
	for _, tt := range []struct {
		name, proxy, token    string
		disablePOToken        bool
		wantProxy, wantPOArgs string
	}{
		{name: "no proxy or token configured"},
		{name: "ignore blank proxy", proxy: " \t "},
		{name: "proxy is forwarded", proxy: proxyURL, wantProxy: proxyURL},
		{
			name: "enabled token is forwarded", token: testToken,
			wantPOArgs: "youtube:player-client=web;po_token=web+test-token",
		},
		{name: "disabled token is not forwarded", token: testToken, disablePOToken: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ProxyURL, POToken, DisablePOToken = tt.proxy, tt.token, tt.disablePOToken
			args := buildArgs(videoURL, "clip.opus", "90", "5")

			assert.Contains(t, args, videoURL)
			assert.Contains(t, args, "--extract-audio")
			assert.Contains(t, args, "--no-playlist")
			assertOption(t, args, "--output", "clip.opus")
			downloader, ok := optionValue(t, args, "--downloader-args")
			require.True(t, ok, "missing ffmpeg clip arguments")
			clipArgs, ok := strings.CutPrefix(downloader, "ffmpeg:")
			require.True(t, ok, "clip arguments should target ffmpeg")
			assertOption(t, strings.Fields(clipArgs), "-ss", "90")
			assertOption(t, strings.Fields(clipArgs), "-t", "5")
			assertOption(t, args, "--cookies", "cookies.txt")
			assertOption(t, args, "--proxy", tt.wantProxy)
			assertOption(t, args, "--extractor-args", tt.wantPOArgs)
			if tt.disablePOToken {
				assert.NotContains(t, strings.Join(args, "\n"), testToken)
			}
		})
	}
}

func assertOption(t *testing.T, args []string, option, want string) {
	t.Helper()
	value, found := optionValue(t, args, option)
	if want == "" {
		assert.False(t, found, "unexpected %s option", option)
		return
	}
	require.True(t, found, "missing %s option", option)
	assert.Equal(t, want, value, "%s option", option)
}

func optionValue(t *testing.T, args []string, option string) (string, bool) {
	t.Helper()
	for i, arg := range args {
		if arg == option {
			require.Less(t, i+1, len(args), "%s has no value", option)
			return args[i+1], true
		}
		if value, ok := strings.CutPrefix(arg, option+"="); ok {
			return value, true
		}
	}
	return "", false
}
