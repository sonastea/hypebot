package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const Schema string = `
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

var (
	DB *sql.DB

	openOnce sync.Once
	openErr  error
)

// GetDBConn opens hypebase.db in the working directory on first use, creating
// it and its schema if needed, and returns the shared connection.
func GetDBConn() (*sql.DB, error) {
	openOnce.Do(func() {
		db, err := sql.Open("sqlite3", "file:hypebase.db?mode=rwc&journal_mode=WAL")
		if err != nil {
			openErr = err
			return
		}

		if _, err = db.Exec(Schema); err != nil {
			db.Close()
			openErr = fmt.Errorf("setting up schema: %w", err)
			return
		}

		DB = db
	})

	return DB, openErr
}
