package themesong

import (
	"database/sql"
	"log"
)

type Themesong struct {
	id       int
	Guild_ID string `json:"guild_id"`
	User_ID  string `json:"user_id"`
	Filepath string `json:"filepath"`
}

var (
	SONG_ADDED_SUCCESS   = "Added themesong ✅"
	SONG_ADDED_FAIL      = "Error adding themesong ⚠️"
	SONG_REMOVED_SUCCESS = "Removed themesong ❌"
	SONG_REMOVED_FAIL    = "Error removing themesong ⚠️"
	SONG_UPDATED_SUCCESS = "Updated themesong ✅"
	SONG_UPDATED_FAIL    = "Error updating themesong ⚠️"
)

func Set(db *sql.DB, file_path string, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("INSERT OR IGNORE INTO Themesong (id, guild_id, user_id, Filepath) VALUES (?, ?, ?, ?);")
	if err != nil {
		log.Println(err)
		return SONG_ADDED_FAIL
	}

	res, err := stmt.Exec(nil, guild_id, user_id, file_path)
	if err != nil {
		log.Println(err)
		return SONG_ADDED_FAIL
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return SONG_ADDED_FAIL
	}
	defer stmt.Close()

	// Check if themesong was added to the database
	if rows > 0 {
		log.Printf("%v for Guild:%v - User:%v \n", SONG_ADDED_SUCCESS, guild_id, user_id)
		return SONG_ADDED_SUCCESS
	}

	return SONG_ADDED_FAIL
}

func Remove(db *sql.DB, guild_id string, user_id string) (message string) {
	stmt, err := db.Prepare("DELETE FROM Themesong WHERE guild_id = ? AND user_id = ?;")
	if err != nil {
		log.Println(err)
		return SONG_REMOVED_FAIL
	}

	res, err := stmt.Exec(guild_id, user_id)
	if err != nil {
		log.Println(err)
		return SONG_REMOVED_FAIL
	}
	defer stmt.Close()

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return SONG_REMOVED_FAIL
	}

	// Check if themesong was updated true
	if rows > 0 {
		log.Printf("%v for Guild:%v - User:%v \n", SONG_REMOVED_SUCCESS, guild_id, user_id)
		return SONG_REMOVED_SUCCESS
	}

	return SONG_REMOVED_FAIL
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
	defer stmt.Close()

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	if rows > 0 {
		log.Printf("%v for Guild:%v - User:%v \n", SONG_UPDATED_SUCCESS, guild_id, user_id)
		return SONG_UPDATED_SUCCESS
	}

	return SONG_UPDATED_FAIL
}
