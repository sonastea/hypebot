package guild

import (
	"log"

	"github.com/sonastea/hypebot/internal/datastore"
)

type Guild struct {
	id        int
	UID       string `json:"uid"`
	VCS       map[string][]string
	Playing   bool
	Active    int8   `json:"active"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type GuildStore interface {
	Add(guild_id string)
	Find(guild_id string) (bool, error)
	Get(guild_id string) *Guild
	Remove(guild_id string)
	GetTotalServed() (uint64, bool)
}

type CacheStore map[string]*Guild

type Store struct {
	datastore.BaseStore
}

var _ GuildStore = &Store{}

func NewGuildCacheStore() CacheStore {
	return make(map[string]*Guild)
}

func NewGuildStore(store datastore.BaseStore) *Store {
	return &Store{BaseStore: store}
}

func (gs *Store) Add(guild_id string) {
	stmt, err := gs.DB().Prepare("INSERT OR IGNORE INTO Guild (UID) VALUES (?);")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	// Check if guild was added because it didn't exist
	if rows > 0 {
		log.Printf("Added Guild:%v to database. \n", guild_id)
	}
}

func (gs *Store) Find(guild_id string) (bool, error) {
	stmt, err := gs.DB().Prepare("SELECT UID from Guild WHERE Guild.UID = ?;")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	if rows > 0 {
		return true, nil
	}

	return false, nil
}

func (gs *Store) Get(guild_id string) *Guild {
	res, err := gs.DB().Query("SELECT UID, Active, CreatedAt, UpdatedAt from Guild Where Guild.UID = ?;", guild_id)
	if err != nil {
		log.Println(err)
	}
	defer res.Close()

	var guild = &Guild{
		VCS: make(map[string][]string),
	}

	for res.Next() {
		err = res.Scan(&guild.UID, &guild.Active, &guild.CreatedAt, &guild.UpdatedAt)
		if err != nil {
			log.Println(err)
		}
	}

	return guild
}

func (gs *Store) Remove(guild_id string) {
	stmt, err := gs.DB().Prepare("UPDATE Guild SET active = 0 WHERE Guild.UID = ?;")
	if err != nil {
		log.Println(err)
	}

	res, err := stmt.Exec(guild_id)
	if err != nil {
		log.Println(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}

	defer stmt.Close()

	if rows > 0 {
		log.Printf("GuildID:%v is now inactive. \n", guild_id)
	}
}

func (gs *Store) GetTotalServed() (uint64, bool) {
	var totalServers uint64

	err := gs.DB().QueryRow("SELECT COUNT(*) FROM Guild;").Scan(&totalServers)
	switch {
	case err != nil:
		log.Println(err)
		return 0, false
	default:
		return totalServers, true
	}
}
