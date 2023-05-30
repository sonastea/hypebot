package models

import (
	"database/sql"
	"time"
)

type User struct {
	id        int
	UID       string     `json:"uid"`
	Guild_ID  string     `json:"guild_id"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type UserStore interface {
	FindUser(db *sql.DB, guild_id string, user_id string) bool
	AddUser(db *sql.DB, user User)
	GetThemesong(db *sql.DB, guild_id string, user_id string) (filePath string, ok bool)
	GetTotalServed(db *sql.DB) (uint64, bool)
}
