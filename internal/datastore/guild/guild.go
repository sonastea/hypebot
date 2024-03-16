package guild

import (
	"database/sql"
	"log"
)

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

type CacheStore map[string]*Guild
type Store struct{}

var _ GuildStore = &Store{}

func NewGuildCacheStore() CacheStore {
	return make(map[string]*Guild)
}

func NewGuildStore() *Store {
	return new(Store)
}

func (gs *Store) AddGuild(db *sql.DB, guild_id string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Guild (UID) VALUES (?);")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	// Check if guild was added because it didn't exist
	if rows > 0 {
		log.Printf("Added Guild:%v to database. \n", guild_id)
	}
}

func (gs *Store) FindGuild(db *sql.DB, guild_id string) (bool, error) {
	stmt, err := db.Prepare("SELECT UID from Guild WHERE Guild.UID = ?;")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (gs *Store) GetGuild(db *sql.DB, guild_id string) *Guild {
	res, err := db.Query("SELECT UID, Active, CreatedAt, UpdatedAt from Guild Where Guild.UID = ?;", guild_id)
	if err != nil {
		log.Println(err)
	}
	defer res.Close()

	var guild = &Guild{
		VCS: make(map[string][]string),
	}

	for res.Next() {
		err = res.Scan(&guild.UID, &guild.Active, &guild.CreatedAt, &guild.UpdatedAt)
		if err != nil {
			log.Println(err)
		}
	}

	return guild
}

func (gs *Store) RemoveGuild(db *sql.DB, guild_id string) {
	stmt, err := db.Prepare("UPDATE Guild SET active = 0 WHERE Guild.UID = ?;")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	if rows > 0 {
		log.Printf("GuildID:%v is now inactive. \n", guild_id)
	}
}

func (gs *Store) GetTotalServed(db *sql.DB) (uint64, bool) {
	var totalServers uint64

	err := db.QueryRow("SELECT COUNT(*) FROM Guild;").Scan(&totalServers)
	switch {
	case err != nil:
		log.Println(err)
		return 0, false
	default:
		return totalServers, true
	}
}
