package guild

import (
	"testing"

	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuildMembership(t *testing.T) {
	gs := NewGuildStore(testutils.NewStore(t))
	gs.Add("guild-a")
	gs.Add("guild-a")
	gs.Add("guild-b")

	g := gs.Get("guild-a")
	require.NotNil(t, g)
	assert.Equal(t, "guild-a", g.UID)
	assert.Equal(t, int8(1), g.Active)
	assert.NotEmpty(t, g.CreatedAt)
	assert.NotEmpty(t, g.UpdatedAt)

	gs.Remove("guild-a")
	removed := gs.Get("guild-a")
	require.NotNil(t, removed)
	assert.Equal(t, "guild-a", removed.UID, "leaving keeps guild history")
	assert.Equal(t, int8(0), removed.Active)

	other := gs.Get("guild-b")
	require.NotNil(t, other)
	assert.Equal(t, "guild-b", other.UID)
	assert.Equal(t, int8(1), other.Active, "leaving does not deactivate other guilds")

	total, ok := gs.GetTotalServed()
	require.True(t, ok)
	assert.Equal(t, uint64(2), total, "repeat joins don't increase the count")
}

func TestGetReturnsNilWhenQueryFails(t *testing.T) {
	store := testutils.NewStore(t)
	_, err := store.DB().Exec("DROP TABLE Guild")
	require.NoError(t, err)

	assert.Nil(t, NewGuildStore(store).Get("guild-a"))
}
