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

	b.Run()
}
