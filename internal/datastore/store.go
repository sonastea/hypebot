package datastore

import (
	"database/sql"
)

type BaseStore interface {
	DB() *sql.DB
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) (*Store, error) {
	return &Store{db: db}, nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}
