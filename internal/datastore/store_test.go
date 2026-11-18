package datastore

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestNew_DBNil(t *testing.T) {
	store, err := New(nil)

	require.EqualError(t, err, "database is nil")
	require.Nil(t, store)
}

func TestNew_DBConnectionFailure(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	store, err := New(db)
	require.EqualError(t, err, "failed to ping database")
	require.Nil(t, store)
}

func TestNew_Success(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	store, err := New(db)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Same(t, db, store.DB())
}
