package database

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetConn forgets the shared connection so each test run opens its own.
func resetConn(t *testing.T) {
	t.Helper()
	reset := func() {
		if DB != nil {
			DB.Close()
		}
		DB, openErr, openOnce = nil, nil, sync.Once{}
	}
	reset()
	t.Cleanup(reset)
}

func TestGetDBConn(t *testing.T) {
	resetConn(t)
	t.Chdir(t.TempDir())
	assert.NoFileExists(t, "hypebase.db", "importing the package must not create the database")

	db, err := GetDBConn()
	require.NoError(t, err)
	require.NotNil(t, db)
	assert.FileExists(t, "hypebase.db")

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM Guild").Scan(&count), "schema is created")

	again, err := GetDBConn()
	require.NoError(t, err)
	assert.Same(t, db, again, "the connection is shared")
}
