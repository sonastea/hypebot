package datastore

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sonastea/hypebot/internal/utils"
)

const create string = `
    CREATE TABLE IF NOT EXISTS User (
    id INTEGER NOT NULL PRIMARY KEY,
    UID STRING NOT NULL UNIQUE
    );`

func NewDBConn() (db *sql.DB, err error) {
	// Create a new database connection
	conn, err := sql.Open("sqlite3", "hypebase.db")
	utils.CheckErr(err)

	// Setup schema
	if _, err := conn.Exec(create); err != nil {
		return nil, err
	}

	// Return a new database connection
	return conn, nil
}
