package hypeserver

import (
	"database/sql"

	"github.com/sonastea/hypebot/internal/pkg/database"
	"github.com/sonastea/hypebot/internal/pkg/datastore/guild"
	"github.com/sonastea/hypebot/internal/pkg/datastore/user"
	"github.com/sonastea/hypebot/internal/utils"
)

var db *sql.DB
var TotalGuilds, TotalUsers uint64

func init() {
	var err error
	db, err = database.GetDBConn()
	utils.CheckErrFatal(err)
}

func GetTotalUsers() uint64 {
    users, success := user.GetTotalServed(db)
	if success != true {
        return TotalUsers
	}
    return users
}

func GetTotalGuilds() uint64 {
    guilds, success := guild.GetTotalServed(db)
	if success != true {
        return TotalGuilds
	}
    return guilds
}

func GetStats() (uint64, uint64) {
    TotalGuilds = GetTotalGuilds()
    TotalUsers = GetTotalUsers()

	return TotalGuilds, TotalUsers
}
