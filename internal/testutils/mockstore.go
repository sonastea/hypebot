package testutils

import "database/sql"

type MockStore struct {
	db *sql.DB
}

func (ms *MockStore) DB() *sql.DB {
	return ms.db
}

func NewMockStore(db *sql.DB) *MockStore {
	return &MockStore{db: db}
}
