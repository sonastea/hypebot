package main

import (
	"github.com/sonastea/hypebot/internal/pkg/hypebot"
	"github.com/sonastea/hypebot/internal/utils"
)

func main() {
	b, err := hypebot.NewHypeBot()
	utils.CheckErr(err)

    b.Run()
}
