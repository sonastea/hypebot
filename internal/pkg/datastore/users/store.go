package users

import (
	"database/sql"
	"fmt"

	"github.com/sonastea/hypebot/internal/utils"
)

func AddUser(db *sql.DB, user User) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO User (id, UID) VALUES (?, ?)")
	utils.CheckQueryErr(err)

	res, err := stmt.Exec(nil, user.UID)
	utils.CheckQueryErr(err)

	rows, err := res.RowsAffected()
	utils.CheckQueryErr(err)

	defer stmt.Close()

	// Check if user was added because it didn't exist
	if rows > 0 {
		fmt.Printf("Added %v \n", user.UID)
	}
}
