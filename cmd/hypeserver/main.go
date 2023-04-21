package main

import (
	"log"

	"github.com/sonastea/hypebot/internal/hypeserver"
)

func main() {
	hs, err := hypeserver.NewHypeServer()
	if err != nil {
		log.Fatalln(err)
	}

	hs.Run()
}
