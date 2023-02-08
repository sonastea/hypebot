package guilds

import (
	"database/sql"
	"log"

	"github.com/sonastea/hypebot/internal/utils"
)

type GuildStore map[string]*Guild

func NewGuildStore() GuildStore {
	return make(map[string]*Guild)
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

func GetGuild(db *sql.DB, guild_id string) *Guild {
	res, err := db.Query("SELECT * from Guild Where Guild.UID = ?;", guild_id)
	utils.CheckErr(err)
	defer res.Close()

	var guild = &Guild{}

	for res.Next() {
		err = res.Scan(&guild.id, &guild.UID, &guild.Active, &guild.CreatedAt, &guild.UpdatedAt)
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
