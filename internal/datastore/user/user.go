package user

import (
	"log"
	"time"

	"github.com/sonastea/hypebot/internal/datastore"
)

type User struct {
	id        int
	UID       string     `json:"uid"`
	Guild_ID  string     `json:"guild_id"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type UserStore interface {
	Find(guild_id string, user_id string) bool
	Add(user User)
	GetThemesong(guild_id string, user_id string) (filePath string, ok bool)
	GetTotalServed() (uint64, bool)
}

type Store struct {
	*datastore.Store
}

var _ UserStore = &Store{}

func NewUserStore(store *datastore.Store) *Store {
	return &Store{Store: store}
}

func (us *Store) Find(guild_id string, user_id string) bool {
	res := us.DB().QueryRow("SELECT UID from User WHERE guild_id = ? AND UID = ?;",
		guild_id, user_id).Scan(&user_id)
	if res != nil {
		return false
	}

	return true
}

func (us *Store) Add(user User) {
	stmt, err := us.DB().Prepare("INSERT OR IGNORE INTO User (guild_id, UID) VALUES (?,?);")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(user.Guild_ID, user.UID)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	// Check if user was added because it didn't exist
	if rows > 0 {
		log.Printf("Added User:%v - Guild:%v \n", user.UID, user.Guild_ID)
	}
}

func (us *Store) GetThemesong(guild_id string, user_id string) (filePath string, ok bool) {
	res, err := us.DB().Query("SELECT Filepath from Themesong Where Themesong.guild_id = ? AND Themesong.user_id = ?;",
		guild_id, user_id)
	if err != nil {
		log.Println(err)
	}
	defer res.Close()

	var filepath string

	for res.Next() {
		err = res.Scan(&filepath)
		if err != nil {
			log.Println(err)
		}
		return filepath, true
	}

	return "", false
}

func (us *Store) GetTotalServed() (uint64, bool) {
	var totalUsers uint64

	err := us.DB().QueryRow("SELECT COUNT(*) FROM User;").Scan(&totalUsers)
	switch {
	case err != nil:
		log.Println(err)
		return 0, false
	default:
		return totalUsers, true
	}
}
