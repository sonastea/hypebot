package user

import (
	"testing"

	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRegisteredUser(t *testing.T) {
	us := NewUserStore(testutils.NewStore(t))
	assert.False(t, us.Find("guild-a", "user-a"), "user starts unregistered")

	us.Add(User{Guild_ID: "guild-a", UID: "user-a"})

	assert.True(t, us.Find("guild-a", "user-a"))
	assert.False(t, us.Find("guild-b", "user-a"), "same user in a different guild")
	assert.False(t, us.Find("guild-a", "user-b"), "different user in the same guild")
}

func TestGetThemesongNotFound(t *testing.T) {
	store := testutils.NewStore(t)
	us := NewUserStore(store)

	path, ok := us.GetThemesong("guild-a", "user-a")
	assert.False(t, ok, "no song set")
	assert.Empty(t, path)

	_, err := store.DB().Exec(`INSERT INTO Themesong (guild_id, user_id, Filepath) VALUES ('guild-a', 'user-a', NULL)`)
	require.NoError(t, err)
	path, ok = us.GetThemesong("guild-a", "user-a")
	assert.False(t, ok, "a row without a file is not a playable song")
	assert.Empty(t, path)

	_, err = store.DB().Exec("DROP TABLE Themesong")
	require.NoError(t, err)
	path, ok = us.GetThemesong("guild-a", "user-a")
	assert.False(t, ok, "query errors are reported as not found")
	assert.Empty(t, path)
}
