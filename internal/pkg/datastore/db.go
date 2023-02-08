package datastore

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sonastea/hypebot/internal/utils"
)

const create string = `
    CREATE TABLE IF NOT EXISTS Guild (
    "id" INTEGER,
    "UID" TEXT NOT NULL UNIQUE,
    "active" INTEGER DEFAULT 1,
    "createdAt"	TEXT DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"	TEXT DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "Guild_PK" PRIMARY KEY("id" AUTOINCREMENT)
    );

    CREATE TABLE IF NOT EXISTS User (
    "id" INTEGER NOT NULL,
    "UID" TEXT NOT NULL,
    "guild_id" TEXT NOT NULL,
    "createdAt"	TEXT DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"	TEXT DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "User_PK" PRIMARY KEY("id" AUTOINCREMENT)
    );

    CREATE TABLE IF NOT EXISTS Themesong (
    "id"       INTEGER NOT NULL,
    "guild_id" TEXT NOT NULL,
    "user_id"  TEXT NOT NULL,
    "Filepath" TEXT,
    CONSTRAINT "Guild_PK" PRIMARY KEY("id" AUTOINCREMENT)
    FOREIGN KEY ("Guild_ID") REFERENCES Guild(UID)
    FOREIGN KEY ("User_ID") REFERENCES User(UID)
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
