package database

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

var conn *sql.DB

func init() {
    var err error

    conn, err = sql.Open("sqlite3", "file:hypebase.db?mode=rwc&journal_mode=WAL")
	utils.CheckErr(err)

	// Setup schema
	if _, err := conn.Exec(create); err != nil {
        utils.CheckErrFatal(err)
	}
}

func GetDBConn() (db *sql.DB, err error) {
	return conn, nil
}
