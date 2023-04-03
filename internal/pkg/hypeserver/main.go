package hypeserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sonastea/hypebot/internal/pkg/hypebot"
)

type hypeserver struct {
	server  *http.Server
	servers uint64
	users   uint64
}

var s hypeserver

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", stats)

	s.server = &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
}

func Run() {
	log.Println("HypeServer listening on port 3000")
	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %s", err)
		}
	}()

	cleanup := make(chan os.Signal, 1)
	signal.Notify(cleanup, os.Interrupt, syscall.SIGINT)
	<-cleanup

	go func() {
		<-cleanup
	}()

	cleansedCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := s.server.Shutdown(cleansedCtx); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	} else {
		log.Printf("Shutdown successful\n")
	}
}

func stats(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)
	s.servers, s.users = hypebot.GetStats()

	data["servers"] = strconv.FormatUint(s.servers, 10)
	data["users"] = strconv.FormatUint(s.users, 10)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)

	return
}
