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
)

type HypeServer struct {
	server  *http.Server
	servers uint64
	users   uint64
}

var totalServers, totalUsers uint64

func NewHypeServer() *HypeServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", stats)

	s := &HypeServer{
		server: &http.Server{
			Addr:    ":3000",
			Handler: mux,
		},
	}

	return s
}

func (hs *HypeServer) Run() {
	log.Println("HypeServer listening on port 3000")
	go func() {
		if err := hs.server.ListenAndServe(); err != http.ErrServerClosed {
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

	if err := hs.server.Shutdown(cleansedCtx); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	} else {
		log.Printf("Shutdown successful\n")
	}
}

func stats(w http.ResponseWriter, r *http.Request) {
	enableCors(&w, r)

	data := make(map[string]string)
	totalServers, totalUsers = GetStats()

	data["servers"] = strconv.FormatUint(totalServers, 10)
	data["users"] = strconv.FormatUint(totalUsers, 10)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)

	return
}

func enableCors(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
}
