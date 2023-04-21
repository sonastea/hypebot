package guild

import (
	"database/sql"
	"log"

	"github.com/sonastea/hypebot/internal/hypebot/models"
	"github.com/sonastea/hypebot/internal/utils"
)

type Store map[string]*models.Guild

func NewGuildStore() Store {
	return make(map[string]*models.Guild)
}

func AddGuild(db *sql.DB, guild_id string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Guild (UID) VALUES (?);")
	utils.CheckErr(err)

	res, err := stmt.Exec(guild_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if guild was added because it didn't exist
	if rows > 0 {
		log.Printf("Added Guild:%v to database. \n", guild_id)
	}
}

func FindGuild(db *sql.DB, guild_id string) (bool, error) {
	stmt, err := db.Prepare("SELECT UID from Guild WHERE Guild.UID = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(guild_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func GetGuild(db *sql.DB, guild_id string) *models.Guild {
	res, err := db.Query("SELECT UID, Active, CreatedAt, UpdatedAt from Guild Where Guild.UID = ?;", guild_id)
	utils.CheckErr(err)
	defer res.Close()

	var guild = &models.Guild{
		VCS: make(map[string][]string),
	}

	for res.Next() {
		err = res.Scan(&guild.UID, &guild.Active, &guild.CreatedAt, &guild.UpdatedAt)
		utils.CheckErr(err)
	}

	return guild
}

func RemoveGuild(db *sql.DB, guild_id string) {
	stmt, err := db.Prepare("UPDATE Guild SET active = 0 WHERE Guild.UID = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(guild_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	if rows > 0 {
		log.Printf("GuildID:%v is now inactive. \n", guild_id)
	}
}

func GetTotalServed(db *sql.DB) (uint64, bool) {
	var totalUsers uint64

	err := db.QueryRow("SELECT COUNT(*) FROM Guild;").Scan(&totalUsers)
	switch {
	case err != nil:
		utils.CheckErr(err)
		return 0, false
	default:
		return totalUsers, true
	}
}
