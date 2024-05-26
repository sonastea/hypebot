package themesong

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var (
	db   *sql.DB
	err  error
	mock sqlmock.Sqlmock
	song Themesong
)

func TestMain(m *testing.M) {
	db, mock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("error with creating sqlmock: %v", err)
	}
	defer db.Close()

	song = Themesong{
		id:       9,
		Guild_ID: "9",
		User_ID:  "9",
		Filepath: "/path/to/themesong",
	}

	os.Exit(m.Run())
}

func TestSet(t *testing.T) {
	mock.ExpectPrepare(`INSERT OR IGNORE INTO Themesong \(id, guild_id, user_id, Filepath\) VALUES \(\?, \?, \?, \?\);`).
		ExpectExec().
		WithArgs(nil, song.Guild_ID, song.User_ID, song.Filepath).
		WillReturnResult(sqlmock.NewResult(9, 1))

	err := Set(db, song.Filepath, song.Guild_ID, song.User_ID)

	assert.Contains(t, SONG_ADDED_SUCCESS, err, SONG_ADDED_FAIL)
}

func TestRemove(t *testing.T) {
	mock.ExpectPrepare(`DELETE FROM Themesong WHERE guild_id = \? AND user_id = \?;`).
		ExpectExec().
		WithArgs(song.Guild_ID, song.User_ID).
		WillReturnResult(sqlmock.NewResult(9, 1))

	err := Remove(db, song.Guild_ID, song.User_ID)

	assert.Contains(t, SONG_REMOVED_SUCCESS, err, SONG_REMOVED_FAIL)
}

func TestUpdate(t *testing.T) {
	mock.ExpectPrepare(`UPDATE Themesong SET Filepath = \? WHERE guild_id = \? AND user_id = \?;`).
		ExpectExec().
		WithArgs(song.Filepath, song.Guild_ID, song.User_ID).
		WillReturnResult(sqlmock.NewResult(9, 1))

	err := Update(db, song.Filepath, song.Guild_ID, song.User_ID)

	assert.Contains(t, SONG_UPDATED_SUCCESS, err, SONG_UPDATED_FAIL)
}
