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
