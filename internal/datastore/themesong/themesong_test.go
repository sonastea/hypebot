package themesong_test

import (
	"testing"

	"github.com/sonastea/hypebot/internal/datastore/themesong"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemesongLifecycle(t *testing.T) {
	store := testutils.NewStore(t)
	db := store.DB()
	us := user.NewUserStore(store)

	checkSong := func(guild, user, want string) {
		t.Helper()
		path, found := us.GetThemesong(guild, user)
		assert.Equal(t, want != "", found, "song exists for %s/%s", guild, user)
		assert.Equal(t, want, path, "song path for %s/%s", guild, user)
	}

	require.Equal(t, "Added themesong ✅", themesong.Set(db, "original.dca", "guild-a", "user-a"))
	require.Equal(t, "Added themesong ✅", themesong.Set(db, "other-user.dca", "guild-a", "user-b"))
	require.Equal(t, "Added themesong ✅", themesong.Set(db, "other-guild.dca", "guild-b", "user-a"))
	checkSong("guild-a", "user-a", "original.dca")

	require.Equal(t, "Updated themesong ✅", themesong.Update(db, "replacement.dca", "guild-a", "user-a"))
	checkSong("guild-a", "user-a", "replacement.dca")
	checkSong("guild-a", "user-b", "other-user.dca")
	checkSong("guild-b", "user-a", "other-guild.dca")

	require.Equal(t, "Removed themesong ❌", themesong.Remove(db, "guild-a", "user-a"))
	checkSong("guild-a", "user-a", "")
	checkSong("guild-a", "user-b", "other-user.dca")
	checkSong("guild-b", "user-a", "other-guild.dca")

	assert.Equal(t, "Error updating themesong ⚠️", themesong.Update(db, "missing.dca", "guild-a", "user-a"))
	assert.Equal(t, "Error removing themesong ⚠️", themesong.Remove(db, "guild-a", "user-a"))
}
