package datastore

import (
	"database/sql"
	"errors"
	"log"
)

var (
	ERROR_DB_NIL           = "database is nil"
	ERROR_CONNECTION_ERROR = "failed to ping database"
)

type BaseStore interface {
	DB() *sql.DB
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New(ERROR_DB_NIL)
	}

	err := db.Ping()
	if err != nil {
    log.Printf("%v: %v", ERROR_CONNECTION_ERROR, err)
		return nil, errors.New(ERROR_CONNECTION_ERROR)
	}

	return &Store{db: db}, nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}
