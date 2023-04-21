package main

import (
	"github.com/sonastea/hypebot/internal/hypeserver"
)

func main() {
	hs := hypeserver.NewHypeServer()
	hs.Run()
}
