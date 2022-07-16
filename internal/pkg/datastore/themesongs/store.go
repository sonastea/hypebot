package themesongs

import (
	"database/sql"
	"fmt"

	"github.com/sonastea/hypebot/internal/utils"
)

func SetThemesong(db *sql.DB, User_ID string, filepath string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Themesong (id, User_ID, Filepath) VALUES (?, ?, ?)")
	utils.CheckErr(err)

	res, err := stmt.Exec(nil, User_ID, filepath)
	utils.CheckErr(err)

	rows, err := res.RowsAffected()
	utils.CheckErr(err)

	defer stmt.Close()

	// Check if themesong was added to the database
	if rows > 0 {
		fmt.Printf("Added %v \n", filepath)
	}
}
