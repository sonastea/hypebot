package testutils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNewMockStore(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a mock database: %v", err)
	}
	defer db.Close()

	store := NewMockStore(db)
	assert.IsType(t, &MockStore{}, store)
}

func TestMockStoreDB(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a mock database: %v", err)
	}
	defer db.Close()

	store := NewMockStore(db)
	assert.IsType(t, &sql.DB{}, store.DB())
}
