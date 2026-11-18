package hypeserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
)

type HypeServer struct {
	server  *http.Server
	servers uint64
	users   uint64
}

func NewHypeServer(gs *guild.Store, us *user.Store) (*HypeServer, error) {
	mux := http.NewServeMux()
	mux.Handle("/stats", stats(gs, us))

	s := &HypeServer{
		server: &http.Server{
			Addr:    ":3000",
			Handler: mux,
		},
	}

	return s, nil
}

// shutdownTimeout bounds how long Stop waits for in-flight requests.
const shutdownTimeout = 5 * time.Second

func (hs *HypeServer) Run() chan os.Signal {
	log.Println("HypeServer listening on port 3000")
	go func() {
		if err := hs.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server: %s", err)
		}
	}()

	cleanup := make(chan os.Signal, 1)
	signal.Notify(cleanup, os.Interrupt, syscall.SIGINT)
	return cleanup
}

func (hs *HypeServer) Stop(sig chan os.Signal) error {
	signal.Stop(sig)
	close(sig)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := hs.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	log.Printf("Shutdown successful\n")

	return nil
}

func stats(gs *guild.Store, us *user.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)

		// Request-local: handlers run concurrently.
		servers, _ := gs.GetTotalServed()
		users, _ := us.GetTotalServed()

		data := map[string]string{
			"servers": strconv.FormatUint(servers, 10),
			"users":   strconv.FormatUint(users, 10),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	})
}

func enableCors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
}
