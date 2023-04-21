package user

import (
	"database/sql"
	"log"

	"github.com/sonastea/hypebot/internal/hypebot/models"
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
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(user.Guild_ID, user.UID)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	// Check if user was added because it didn't exist
	if rows > 0 {
		log.Printf("Added User:%v - Guild:%v \n", user.UID, user.Guild_ID)
	}
}

func GetThemesong(db *sql.DB, guild_id string, user_id string) (filePath string, ok bool) {
	res, err := db.Query("SELECT Filepath from Themesong Where Themesong.guild_id = ? AND Themesong.user_id = ?;",
		guild_id, user_id)
	if err != nil {
		log.Println(err)
	}
	defer res.Close()

	var filepath string

	for res.Next() {
		err = res.Scan(&filepath)
		if err != nil {
			log.Println(err)
		}
		return filepath, true
	}

	return "", false
}

func GetTotalServed(db *sql.DB) (uint64, bool) {
	var totalUsers uint64

	err := db.QueryRow("SELECT COUNT(*) FROM User;").Scan(&totalUsers)
	switch {
	case err != nil:
		log.Println(err)
		return 0, false
	default:
		return totalUsers, true
	}
}
