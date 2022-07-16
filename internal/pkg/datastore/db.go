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
    );

    CREATE TABLE IF NOT EXISTS Themesong (
	id       INTEGER NOT NULL PRIMARY KEY,
	User_ID  STRING NOT NULL UNIQUE,
	Filepath STRING,
    FOREIGN KEY (User_ID) REFERENCES User(UID)
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
