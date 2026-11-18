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

	srvChan := hs.Run()

	// Stop closes the channel, so it must run exactly once even if more
	// interrupts were buffered.
	<-srvChan
	if err := hs.Stop(srvChan); err != nil {
		log.Fatalln(err)
	}
}
