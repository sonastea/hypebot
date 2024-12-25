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
			if strings.ContainsAny(arg, expected) {
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

	t.Run("POToken is set", func(t *testing.T) {
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, "cookies.txt"), "Missing cookies.txt argument, got %+v", args)
		assert.True(t, containsArgs(args, POToken), "Missing text-token argument, got %+v", args)
	})

	t.Run("POToken is not set", func(t *testing.T) {
		POToken = ""
		args := buildArgs("", "", "", "")

		assert.False(t, !containsArgs(args, "cookies.txt"), "cookies.txt argument should not be present, got %+v", args)
	})

	t.Run("PROXY_URL is set", func(t *testing.T) {
		t.Setenv("PROXY_URL", expectedProxyURL)
		args := buildArgs("", "", "", "")

		assert.True(t, containsArgs(args, expectedProxyURL), "Missing proxy argument, got %+v", args)
	})

	t.Run("PROXY_URL is not set", func(t *testing.T) {
		args := buildArgs("", "", "", "")

		assert.False(t, !containsArgs(args, expectedProxyURL), "Missing proxy argument, got %+v", args)
	})
}
