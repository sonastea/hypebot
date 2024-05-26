package user

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sonastea/hypebot/internal/datastore"
	"github.com/stretchr/testify/assert"
)

var (
	db        *sql.DB
	err       error
	mock      sqlmock.Sqlmock
	store     datastore.BaseStore
	userStore *Store
	user      User
)

func TestMain(m *testing.M) {
	db, mock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("error with creating sqlmock: %v", err)
	}
	defer db.Close()

	store, _ = datastore.New(db)
	userStore = NewUserStore(store)
	user = User{
		Guild_ID: "123",
		UID:      "9",
	}

	os.Exit(m.Run())
}

func TestFind(t *testing.T) {
	columns := []string{"UID"}
	mock.ExpectQuery(`SELECT UID from User WHERE guild_id = \? AND UID = \?;`).
		WithArgs(user.Guild_ID, user.UID).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(user.UID))

	found := userStore.Find(user.Guild_ID, user.UID)
	assert.True(t, found)
}

func TestAdd(t *testing.T) {
	mock.ExpectPrepare("INSERT OR IGNORE INTO User").
		ExpectExec().
		WithArgs(user.Guild_ID, user.UID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	userStore.Add(user)
}

func TestGetThemesong(t *testing.T) {
	expectedFilePath := "/path/to/themesong"

	mock.ExpectQuery(`SELECT Filepath from Themesong Where Themesong.guild_id = \? AND Themesong.user_id = \?;`).
		WithArgs(user.Guild_ID, user.UID).
		WillReturnRows(sqlmock.NewRows([]string{"Filepath"}).AddRow(expectedFilePath))

	filePath, ok := userStore.GetThemesong(user.Guild_ID, user.UID)
	assert.True(t, ok)
	assert.Equal(t, expectedFilePath, filePath)
}

func TestGetTotalServed(t *testing.T) {
	expectedTotal := uint64(9)

	columns := []string{"COUNT(*)"}
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM User;").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(expectedTotal))

	total, ok := userStore.GetTotalServed()
	assert.True(t, ok)
	assert.Equal(t, expectedTotal, total)
}
