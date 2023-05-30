package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDBConn(t *testing.T) {
	t.Run("Get db conn", func(t *testing.T) {
		db, err := GetDBConn()
		assert.Equal(t, db, DB)
		assert.Equal(t, nil, err)
	})
}
