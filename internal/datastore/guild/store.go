package guild

import (
	"database/sql"
	"log"

	"github.com/sonastea/hypebot/internal/hypebot/models"
)

type CacheStore map[string]*models.Guild
type Store struct{}

func NewGuildCacheStore() CacheStore {
	return make(map[string]*models.Guild)
}

func NewGuildStore() *Store {
	return new(Store)
}

func (gs Store) AddGuild(db *sql.DB, guild_id string) {
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

func (gs Store) FindGuild(db *sql.DB, guild_id string) (bool, error) {
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

func (gs Store) GetGuild(db *sql.DB, guild_id string) *models.Guild {
	res, err := db.Query("SELECT UID, Active, CreatedAt, UpdatedAt from Guild Where Guild.UID = ?;", guild_id)
	if err != nil {
		log.Println(err)
	}
	defer res.Close()

	var guild = &models.Guild{
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

func (gs Store) RemoveGuild(db *sql.DB, guild_id string) {
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

func (gs Store) GetTotalServed(db *sql.DB) (uint64, bool) {
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
