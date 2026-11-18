package testutils

import (
	"database/sql"
	"testing"

	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/datastore"
	"github.com/stretchr/testify/require"
)

// NewStore creates an isolated SQLite store with the application's schema.
func NewStore(t testing.TB) *datastore.Store {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)

	_, err = db.Exec(database.Schema)
	require.NoError(t, err)
	store, err := datastore.New(db)
	require.NoError(t, err)
	return store
}
