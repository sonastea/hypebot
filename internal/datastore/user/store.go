package user

import (
	"database/sql"
	"log"

	"github.com/sonastea/hypebot/internal/hypebot/models"
	"github.com/sonastea/hypebot/internal/utils"
)

func FindUser(db *sql.DB, guild_id string, user_id string) bool {
	res := db.QueryRow("SELECT UID from User WHERE guild_id = ? AND UID = ?;",
		guild_id, user_id).Scan(&user_id)
	if res != nil {
		return false
	}

	return true
}

func AddUser(db *sql.DB, user models.User) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO User (guild_id, UID) VALUES (?,?);")
	utils.CheckErr(err)

	res, err := stmt.Exec(user.Guild_ID, user.UID)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if user was added because it didn't exist
	if rows > 0 {
		log.Printf("Added User:%v - Guild:%v \n", user.UID, user.Guild_ID)
	}
}

func GetThemesong(db *sql.DB, guild_id string, user_id string) (filePath string, ok bool) {
	res, err := db.Query("SELECT Filepath from Themesong Where Themesong.guild_id = ? AND Themesong.user_id = ?;",
		guild_id, user_id)
	utils.CheckErr(err)
	defer res.Close()

	var filepath string

	for res.Next() {
		err = res.Scan(&filepath)
		utils.CheckErr(err)
		return filepath, true
	}

	return "", false
}

func GetTotalServed(db *sql.DB) (uint64, bool) {
	var totalUsers uint64

	err := db.QueryRow("SELECT COUNT(*) FROM User;").Scan(&totalUsers)
	switch {
	case err != nil:
		utils.CheckErr(err)
		return 0, false
	default:
		return totalUsers, true
	}
}
