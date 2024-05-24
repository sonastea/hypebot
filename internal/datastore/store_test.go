package datastore

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNew_DBNil(t *testing.T) {
	_, err := New(nil)

	assert.NotNil(t, err, "Expected error passing a nil db")
	assert.Exactly(t, err.Error(), ERROR_DB_NIL)
}

func TestNew_DBConnectionFailure(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to open a mock database: %v", err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(errors.New(ERROR_CONNECTION_ERROR))

	_, err = New(db)
	if err == nil {
		t.Error("Expected error passing a dead db connection")
	}

	assert.Exactly(t, err.Error(), ERROR_CONNECTION_ERROR)
}

func TestNew_Success(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a mock database: %v", err)
	}
	defer db.Close()

	store, err := New(db)
	assert.IsType(t, &Store{}, store)
}
