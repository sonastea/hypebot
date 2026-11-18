package hypebot

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robrotheram/dca"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeYtdlp stands in for the yt-dlp binary. Its --simulate (validation) and
// download responses are controlled by FAKE_YTDLP_* environment variables, and
// it records one line per invocation.
const fakeYtdlp = `#!/bin/sh
echo "$*" >> "$FAKE_YTDLP_LOG"
for arg; do
	if [ "$arg" = "--simulate" ]; then
		printf '%s' "$FAKE_YTDLP_SIMULATE_STDOUT"
		printf '%s' "$FAKE_YTDLP_SIMULATE_STDERR" >&2
		exit "${FAKE_YTDLP_SIMULATE_EXIT:-0}"
	fi
done
out=""; prev=""
for arg; do
	[ "$prev" = "--output" ] && out="$arg"
	prev="$arg"
done
if [ -n "$FAKE_YTDLP_AUDIO" ]; then
	ffmpeg -nostdin -loglevel error -f lavfi -i "sine=frequency=440:duration=1" -c:a libopus "songs/$out" || exit 1
fi
printf '%s' "$FAKE_YTDLP_JSON"
exit "${FAKE_YTDLP_EXIT:-0}"
`

const testVideoURL = "https://youtu.be/dQw4w9WgXcQ"

// useFakeYtdlp puts the fake yt-dlp first on PATH, runs the test in an empty
// working directory, and returns a function listing the recorded invocations.
func useFakeYtdlp(t *testing.T) (calls func() []string) {
	t.Helper()
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "yt-dlp"), []byte(fakeYtdlp), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	logPath := filepath.Join(bin, "calls.log")
	t.Setenv("FAKE_YTDLP_LOG", logPath)
	// Default validation output: a 3.5 minute video that is not live.
	t.Setenv("FAKE_YTDLP_SIMULATE_STDOUT", "False|212\n")
	t.Chdir(t.TempDir())

	return func() []string {
		data, err := os.ReadFile(logPath)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		require.NoError(t, err)
		return strings.Split(strings.TrimSpace(string(data)), "\n")
	}
}

func newDownloadBot() *HypeBot {
	return &HypeBot{downloadSem: make(chan struct{}, 1)}
}

func TestValidateUrl(t *testing.T) {
	t.Run("accepts a video that passes the filters", func(t *testing.T) {
		calls := useFakeYtdlp(t)
		ok, err := newDownloadBot().validateUrl(testVideoURL)
		require.NoError(t, err)
		assert.True(t, ok)
		require.Len(t, calls(), 1)
		assert.Contains(t, calls()[0], testVideoURL)
	})

	t.Run("treats the max-downloads exit as success", func(t *testing.T) {
		useFakeYtdlp(t)
		t.Setenv("FAKE_YTDLP_SIMULATE_EXIT", "101")
		ok, err := newDownloadBot().validateUrl(testVideoURL)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("accepts a video just under the length limit", func(t *testing.T) {
		useFakeYtdlp(t)
		t.Setenv("FAKE_YTDLP_SIMULATE_STDOUT", "False|599.5\n")
		ok, err := newDownloadBot().validateUrl(testVideoURL)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	for _, tt := range []struct{ name, stdout, wantErr string }{
		{"live stream", "True|NA\n", "Unable to process a live video :warning:"},
		{"video at the length limit", "False|600\n", "Video must not exceed 10 minutes :warning:"},
		{"long video", "False|3600\n", "Video must not exceed 10 minutes :warning:"},
		{"unknown length", "NA|NA\n", "Unable to determine the video's length :warning:"},
		{"no metadata printed", "", "Unable to determine the video's length :warning:"},
	} {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			useFakeYtdlp(t)
			t.Setenv("FAKE_YTDLP_SIMULATE_STDOUT", tt.stdout)
			ok, err := newDownloadBot().validateUrl(testVideoURL)
			assert.False(t, ok)
			assert.EqualError(t, err, tt.wantErr)
		})
	}

	t.Run("reports the yt-dlp error to the user", func(t *testing.T) {
		useFakeYtdlp(t)
		t.Setenv("FAKE_YTDLP_SIMULATE_EXIT", "1")
		t.Setenv("FAKE_YTDLP_SIMULATE_STDERR", "ERROR: [youtube] dQw4w9WgXcQ: Private video\n")
		ok, err := newDownloadBot().validateUrl(testVideoURL)
		assert.False(t, ok)
		assert.EqualError(t, err, "Private video :warning:")
	})

	t.Run("fails cleanly when yt-dlp is not installed", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		ok, err := newDownloadBot().validateUrl(testVideoURL)
		assert.False(t, ok)
		assert.EqualError(t, err, "There was an error processing your command :warning:")
	})
}

func TestDownloadVideoFailures(t *testing.T) {
	t.Run("rejected URL is never downloaded", func(t *testing.T) {
		calls := useFakeYtdlp(t)
		t.Setenv("FAKE_YTDLP_SIMULATE_EXIT", "1")
		t.Setenv("FAKE_YTDLP_SIMULATE_STDERR", "ERROR: [youtube] dQw4w9WgXcQ: Video unavailable\n")

		_, err := newDownloadBot().downloadVideo(testVideoURL, "0", "5")
		assert.EqualError(t, err, "Video unavailable :warning:")
		assert.Len(t, calls(), 1, "only the validation call runs")
		assert.NoDirExists(t, "songs")
	})

	for _, tt := range []struct {
		name, exit, json, start, wantErr string
	}{
		{"download fails", "1", "", "0", "There was an error processing your request ⚠️"},
		{"no metadata returned", "0", "", "0", "Unable to retrieve requested audio ⚠️"},
		{"start time past the end of the video", "0", `{"id":"abc","duration":60}`, "61", "Invalid start time ⚠️"},
		// yt-dlp exits 101 when --max-downloads stops it; the download still counts.
		{"max-downloads exit is not a failure", "101", `{"id":"abc","duration":60}`, "61", "Invalid start time ⚠️"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := useFakeYtdlp(t)
			t.Setenv("FAKE_YTDLP_EXIT", tt.exit)
			t.Setenv("FAKE_YTDLP_JSON", tt.json)

			path, err := newDownloadBot().downloadVideo(testVideoURL, tt.start, "5")
			assert.EqualError(t, err, tt.wantErr)
			assert.Empty(t, path)
			assert.Len(t, calls(), 2, "validation then download")
		})
	}
}

func TestDownloadVideoCreatesPlayableSong(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is required to encode songs; CI installs it")
	}
	useFakeYtdlp(t)
	t.Setenv("FAKE_YTDLP_AUDIO", "1")
	t.Setenv("FAKE_YTDLP_JSON", `{"id":"abc","title":"Test tone","duration":120}`)

	path, err := newDownloadBot().downloadVideo(testVideoURL, "0", "1")
	require.NoError(t, err)
	assert.Regexp(t, `^\./songs/[0-9a-f-]{36}\.dca$`, path)

	entries, err := os.ReadDir("songs")
	require.NoError(t, err)
	require.Len(t, entries, 1, "intermediate .opus files are removed")
	assert.Equal(t, filepath.Base(path), entries[0].Name())

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	decoder := dca.NewDecoder(f)
	frames := 0
	for {
		_, err := decoder.OpusFrame()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		frames++
	}
	assert.Greater(t, frames, 10, "a one-second clip yields many 20ms opus frames")
}
