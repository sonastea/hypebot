package users

import (
	"database/sql"
	"fmt"

	"github.com/sonastea/hypebot/internal/utils"
)

func FindUser(db *sql.DB, user_id string) (bool, error) {
	stmt, err := db.Prepare("SELECT UID from User Where User.UID = ?;")
	utils.CheckErr(err)

	res, err := stmt.Exec(user_id)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func AddUser(db *sql.DB, user User) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO User (UID) VALUES (?);")
	utils.CheckErr(err)

	res, err := stmt.Exec(user.UID)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if user was added because it didn't exist
	if rows > 0 {
		fmt.Printf("Added %v \n", user.UID)
	}
}

func GetThemesong(db *sql.DB, user_id string) (filePath string, ok bool) {
	res, err := db.Query("SELECT Filepath from Themesong Where Themesong.user_id = ?;", user_id)
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
