package hypebot

import (
	"context"
	"database/sql"

	"github.com/sonastea/hypebot/internal/pkg/datastore"
	"github.com/sonastea/hypebot/internal/pkg/datastore/guilds"
	"github.com/sonastea/hypebot/internal/pkg/datastore/users"
	"github.com/sonastea/hypebot/internal/utils"
)

var db *sql.DB
var TotalServers, TotalUsers uint64

func init() {
	var err error
	db, err = datastore.NewDBConn()
	utils.CheckErrFatal(err)
}

func GetTotalUsers() {
	users.GetTotalServed(db)
}

func GetTotalServers() {
	guilds.GetTotalServed(db)
}

func GetStats() (uint64, uint64) {
	err := error(nil)
	ctx := context.Background()

	prev := TotalServers
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Guild;").Scan(&TotalServers)
	if err != nil {
		utils.CheckErr(err)
		TotalServers = prev
	}

	prev = TotalUsers
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Guild;").Scan(&TotalUsers)
	if err != nil {
		utils.CheckErr(err)
		TotalUsers = prev
	}

	return TotalServers, TotalUsers
}
