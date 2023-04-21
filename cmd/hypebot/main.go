package main

import (
	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/hypebot"
	"github.com/sonastea/hypebot/internal/utils"
)

func main() {
    db, err := database.GetDBConn()
    utils.CheckErrFatal(err)

	b, err := hypebot.NewHypeBot(db)
	utils.CheckErrFatal(err)

	b.Run()
}
