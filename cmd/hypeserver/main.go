package main

import (
	"log"

	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/hypeserver"
)

func main() {
	db, err := database.GetDBConn()
	if err != nil {
		log.Fatalln(err)
	}

	hs, err := hypeserver.NewHypeServer(db)
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
