package database

import (
	"database/sql"
	"log"

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

var DB *sql.DB

func init() {
	var err error

	DB, err = sql.Open("sqlite3", "file:hypebase.db?mode=rwc&journal_mode=WAL")
	if err != nil {
		log.Fatalln(err)
	}

	// Setup schema
	if _, err = DB.Exec(Schema); err != nil {
		log.Fatalln(err)
	}
}

func GetDBConn() (db *sql.DB, err error) {
	return DB, nil
}
