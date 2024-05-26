package guild

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
	db         *sql.DB
	err        error
	guild      Guild
	mock       sqlmock.Sqlmock
	store      datastore.BaseStore
	guildStore *Store
)

func TestMain(m *testing.M) {
	db, mock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("error with creating sqlmock: %v", err)
	}
	defer db.Close()

	store, _ = datastore.New(db)
	guildStore = NewGuildStore(store)
	guild = Guild{
		UID:       "9",
		Active:    1,
		CreatedAt: "2009-09-09",
		UpdatedAt: "2009-09-09",
	}

	os.Exit(m.Run())
}

func TestAdd(t *testing.T) {
  mock.ExpectPrepare(`INSERT OR IGNORE INTO Guild \(UID\) VALUES \(\?\);`).
    ExpectExec().
  WithArgs(guild.UID).
  WillReturnResult(sqlmock.NewResult(1, 1))

  guildStore.Add(guild.UID)

  err := mock.ExpectationsWereMet()
  assert.NoError(t, err)
}

func TestFind(t *testing.T) {
	mock.ExpectPrepare(`SELECT UID FROM Guild WHERE Guild.UID = \?;`).
		ExpectExec().
		WithArgs(guild.UID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	found, err := guildStore.Find(guild.UID)

	assert.NoError(t, err)
	assert.True(t, found)
}

func TestGet(t *testing.T) {
	columns := []string{"UID", "Active", "CreatedAt", "UpdatedAt"}
	mock.ExpectQuery(`SELECT UID, Active, CreatedAt, UpdatedAt from Guild Where Guild.UID = \?;`).
		WithArgs(guild.UID).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(guild.UID, guild.Active, guild.CreatedAt, guild.UpdatedAt))

	g := guildStore.Get(guild.UID)

	assert.Equal(t, guild.UID, g.UID)
	assert.Equal(t, guild.Active, g.Active)
	assert.Equal(t, guild.CreatedAt, g.CreatedAt)
	assert.Equal(t, guild.UpdatedAt, g.UpdatedAt)
}

func TestRemove(t *testing.T) {
	mock.ExpectPrepare(`UPDATE Guild SET active = 0 WHERE Guild.UID = \?;`).
		ExpectExec().
		WithArgs(guild.UID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	guildStore.Remove(guild.UID)

	err := mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetTotalServed(t *testing.T) {
	expectedTotal := uint64(9)

	columns := []string{"COUNT(*)"}
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM Guild;").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(expectedTotal))

	total, ok := guildStore.GetTotalServed()

	assert.True(t, ok)
	assert.Equal(t, expectedTotal, total)
}
