package main

import (
	"log"

	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/datastore"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/hypeserver"
)

func main() {
	db, err := database.GetDBConn()
	if err != nil {
		log.Fatalln(err)
	}

  store, err := datastore.New(db)
	if err != nil {
		log.Fatalln(err)
	}

	gs := guild.NewGuildStore(store)
	us := user.NewUserStore(store)

	hs, err := hypeserver.NewHypeServer(gs, us)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, srvChan := hs.Run()

	for range srvChan {
		err = hs.Stop(ctx, srvChan)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
