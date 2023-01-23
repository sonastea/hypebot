package themesongs

import (
	"database/sql"
	"fmt"

	"github.com/sonastea/hypebot/internal/utils"
)

var message string

func SetThemesong(db *sql.DB, file_path string, user_id string) (message string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Themesong (id, user_id, Filepath) VALUES (?, ?, ?);")
	utils.CheckErr(err)

	res, err := stmt.Exec(nil, user_id, file_path)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if themesong was added to the database
	if rows > 0 {
		message := fmt.Sprintf("Added themesong ✅")
		fmt.Println(message)
		return message
	}

	message = "Error adding themesong"
	return message
}

func UpdateThemesong(db *sql.DB, file_path string, user_id string) (message string) {
	stmt, err := db.Prepare("UPDATE Themesong SET Filepath = ? WHERE user_id = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(file_path, user_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	if rows > 0 {
		message := fmt.Sprintf("Updated themesong ✅")
		fmt.Println(message)
		return message
	}

	message = "Error updating themesong"
	return message
}

func RemoveThemesong(db *sql.DB, user_id string) (message string) {
	stmt, err := db.Prepare("DELETE from Themesong WHERE user_id = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(user_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if themesong was updated true
	if rows > 0 {
		message := fmt.Sprintf("Removed themesong ❌")
		fmt.Println(message)
		return message
	}

	message = "Error removing themesong ⚠️"
	return message
}
