package main

import (
	"log"

	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/hypebot"
)

func main() {
	db, err := database.GetDBConn()
	if err != nil {
		log.Fatalln(err)
	}

	b, err := hypebot.NewHypeBot(db)
	if err != nil {
		log.Fatalln(err)
	}

	botChan := b.Run()
	if botChan == nil {
		log.Fatal()
	}

  defer close(botChan)

	select {
	case <-botChan:
    log.Println("Received stop signal, shutting down hypebot...")
		if err := b.Stop(); err != nil {
      log.Fatalf("Error while stopping hypebot: %v", err)
    }
	}
}
