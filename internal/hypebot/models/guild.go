package models

import "database/sql"

type Guild struct {
	id        int
	UID       string `json:"uid"`
	VCS       map[string][]string
	Playing   bool
	Active    int8   `json:"active"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type GuildStore interface {
	AddGuild(db *sql.DB, guild_id string)
	FindGuild(db *sql.DB, guild_id string) (bool, error)
	GetGuild(db *sql.DB, guild_id string) *Guild
	RemoveGuild(db *sql.DB, guild_id string)
	GetTotalServed(db *sql.DB) (uint64, bool)
}
