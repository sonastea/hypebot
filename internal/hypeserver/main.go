package hypeserver

import (
	"context"
	"encoding/json"
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

var TotalServers, TotalUsers uint64

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

func (hs *HypeServer) Run() (context.Context, chan os.Signal) {
	log.Println("HypeServer listening on port 3000")
	go func() {
		if err := hs.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server: %s", err)
		}
	}()

	ctx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	cleanup := make(chan os.Signal, 1)
	signal.Notify(cleanup, os.Interrupt, syscall.SIGINT)
	return ctx, cleanup
}

func (hs *HypeServer) Stop(ctx context.Context, sig chan os.Signal) error {
	close(sig)

	if err := hs.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Shutdown error: %v\n", err)
	}

	log.Printf("Shutdown successful\n")

	return nil
}

func stats(gs *guild.Store, us *user.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w, r)

		data := make(map[string]string)
		TotalServers, _ = gs.GetTotalServed()
		TotalUsers, _ = us.GetTotalServed()

		data["servers"] = strconv.FormatUint(TotalServers, 10)
		data["users"] = strconv.FormatUint(TotalUsers, 10)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	})
}

func enableCors(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
}
