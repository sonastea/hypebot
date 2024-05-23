package themesong

import (
	"database/sql"
	"fmt"
	"log"
)

type Themesong struct {
	id       int
	Guild_ID string `json:"guild_id"`
	User_ID  string `json:"user_id"`
	Filepath string `json:"filepath"`
}

var message string

func Set(db *sql.DB, file_path string, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Themesong (id, guild_id, user_id, Filepath) VALUES (?, ?, ?, ?);")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(nil, guild_id, user_id, file_path)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

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

func Update(db *sql.DB, file_path string, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("UPDATE Themesong SET Filepath = ? WHERE guild_id = ? AND user_id = ?;")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(file_path, guild_id, user_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	if rows > 0 {
		message := fmt.Sprintf("Updated themesong ✅")
		log.Printf("%v for Guild:%v - User:%v \n", message, guild_id, user_id)
		return message
	}

	message = "Error updating themesong"
	return message
}

func Remove(db *sql.DB, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("DELETE from Themesong WHERE guild_id = ? AND user_id = ?;")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id, user_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

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
