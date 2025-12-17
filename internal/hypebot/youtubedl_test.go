package hypebot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func containsArgs(args []string, expectedArgs ...string) bool {
	for _, expected := range expectedArgs {
		found := false
		for _, arg := range args {
			if strings.Contains(arg, expected) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestParseYtdlpError(t *testing.T) {
	t.Run("extracts message after third colon", func(t *testing.T) {
		stderr := `ERROR: [youtube] abc123: Some error message`
		result := parseYtdlpError(stderr)

		assert.Equal(t, "Some error message", result)
	})

	t.Run("extracts message when only two parts", func(t *testing.T) {
		stderr := `ERROR: Something went wrong`
		result := parseYtdlpError(stderr)

		assert.Equal(t, "Something went wrong", result)
	})

	t.Run("returns default message when no ERROR line", func(t *testing.T) {
		stderr := `[youtube] some other output`
		result := parseYtdlpError(stderr)

		assert.Equal(t, "Unable to process video", result)
	})

	t.Run("returns default message for empty stderr", func(t *testing.T) {
		result := parseYtdlpError("")

		assert.Equal(t, "Unable to process video", result)
	})
}

func TestAppendCommonYtdlpArgs(t *testing.T) {
	t.Run("always includes cookies.txt", func(t *testing.T) {
		ProxyURL = ""
		DisablePOToken = true
		args := appendCommonYtdlpArgs([]string{})

		assert.True(t, containsArgs(args, "cookies.txt"), "cookies.txt should always be present")
	})

	t.Run("includes proxy when set", func(t *testing.T) {
		ProxyURL = "http://proxy.example.com"
		DisablePOToken = true
		args := appendCommonYtdlpArgs([]string{})

		assert.True(t, containsArgs(args, "--proxy", "http://proxy.example.com"))
	})

	t.Run("excludes proxy when empty", func(t *testing.T) {
		ProxyURL = ""
		DisablePOToken = true
		args := appendCommonYtdlpArgs([]string{})

		assert.False(t, containsArgs(args, "--proxy"))
	})

	t.Run("excludes proxy when whitespace only", func(t *testing.T) {
		ProxyURL = "   "
		DisablePOToken = true
		args := appendCommonYtdlpArgs([]string{})

		assert.False(t, containsArgs(args, "--proxy"))
	})

	t.Run("includes POToken when enabled and set", func(t *testing.T) {
		ProxyURL = ""
		POToken = "test-po-token"
		DisablePOToken = false
		args := appendCommonYtdlpArgs([]string{})

		assert.True(t, containsArgs(args, "--extractor-args"))
		assert.True(t, containsArgs(args, "test-po-token"))
	})

	t.Run("excludes POToken when disabled", func(t *testing.T) {
		ProxyURL = ""
		POToken = "test-po-token"
		DisablePOToken = true
		args := appendCommonYtdlpArgs([]string{})

		assert.False(t, containsArgs(args, "test-po-token"))
	})

	t.Run("excludes POToken when empty", func(t *testing.T) {
		ProxyURL = ""
		POToken = ""
		DisablePOToken = false
		args := appendCommonYtdlpArgs([]string{})

		assert.False(t, containsArgs(args, "--extractor-args"))
	})

	t.Run("preserves existing args", func(t *testing.T) {
		ProxyURL = ""
		DisablePOToken = true
		existingArgs := []string{"--extract-audio", "--no-playlist"}
		args := appendCommonYtdlpArgs(existingArgs)

		assert.True(t, containsArgs(args, "--extract-audio", "--no-playlist", "cookies.txt"))
	})
}

func TestBuildArgs(t *testing.T) {
	POToken = "test-token"
	expectedProxyURL := "http://PROXYURLVALUE"

	t.Run("cookies.txt is always included", func(t *testing.T) {
		DisablePOToken = true
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, "cookies.txt"), "Missing cookies.txt argument, got %+v", args)
	})

	t.Run("POToken is set and enabled", func(t *testing.T) {
		POToken = "test-token"
		DisablePOToken = false
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, "cookies.txt"), "Missing cookies.txt argument, got %+v", args)
		assert.True(t, containsArgs(args, POToken), "Missing test-token argument, got %+v", args)
	})

	t.Run("POToken is set but disabled", func(t *testing.T) {
		POToken = "test-token"
		DisablePOToken = true
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, "cookies.txt"), "cookies.txt should always be present, got %+v", args)
		assert.False(t, containsArgs(args, POToken), "POToken argument should not be present when disabled, got %+v", args)
	})

	t.Run("POToken is not set", func(t *testing.T) {
		POToken = ""
		DisablePOToken = false
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, "cookies.txt"), "cookies.txt should always be present, got %+v", args)
	})

	t.Run("PROXY_URL is set", func(t *testing.T) {
		ProxyURL = expectedProxyURL
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, expectedProxyURL), "Missing proxy argument, got %+v", args)
	})

	t.Run("PROXY_URL is not set", func(t *testing.T) {
		ProxyURL = ""
		args := buildArgs("", "", "", "")

		assert.False(t, containsArgs(args, expectedProxyURL), "Proxy argument should not be present, got %+v", args)
	})
}
