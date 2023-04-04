package themesongs

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sonastea/hypebot/internal/utils"
)

var message string

func SetThemesong(db *sql.DB, file_path string, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Themesong (id, guild_id, user_id, Filepath) VALUES (?, ?, ?, ?);")
	utils.CheckErr(err)

	res, err := stmt.Exec(nil, guild_id, user_id, file_path)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if themesong was added to the database
	if rows > 0 {
		message := fmt.Sprintf("Added themesong ✅")
		log.Printf("%v for Guild:%v - User:%v \n", message, guild_id, user_id)
		return message
	}

	message = "Error adding themesong"
	return message
}

func UpdateThemesong(db *sql.DB, file_path string, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("UPDATE Themesong SET Filepath = ? WHERE guild_id = ? AND user_id = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(file_path, guild_id, user_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	if rows > 0 {
		message := fmt.Sprintf("Updated themesong ✅")
		log.Printf("%v for Guild:%v - User:%v \n", message, guild_id, user_id)
		return message
	}

	message = "Error updating themesong"
	return message
}

func RemoveThemesong(db *sql.DB, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("DELETE from Themesong WHERE guild_id = ? AND user_id = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(guild_id, user_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if themesong was updated true
	if rows > 0 {
		message := fmt.Sprintf("Removed themesong ❌")
		log.Printf("%v for Guild:%v - User:%v \n", message, guild_id, user_id)
		return message
	}

	message = "Error removing themesong ⚠️"
	return message
}
