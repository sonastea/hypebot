package main

import (
	"github.com/sonastea/hypebot/internal/pkg/hypeserver"
)

func main() {
	hs := hypeserver.NewHypeServer()
	hs.Run()
}
