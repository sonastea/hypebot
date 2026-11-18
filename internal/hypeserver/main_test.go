package hypeserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAndStop(t *testing.T) {
	store := testutils.NewStore(t)
	hs, err := NewHypeServer(guild.NewGuildStore(store), user.NewUserStore(store))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, hs.server.Close()) })

	hs.server.Addr = "127.0.0.1:0"
	listening := make(chan string, 1)
	hs.server.BaseContext = func(listener net.Listener) context.Context {
		listening <- listener.Addr().String()
		return context.Background()
	}

	// A request that stays in flight until released, to observe graceful shutdown.
	started, release := make(chan struct{}), make(chan struct{})
	hs.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		io.WriteString(w, "done")
	})

	srvChan := hs.Run()
	require.NotNil(t, srvChan)
	t.Cleanup(func() { signal.Stop(srvChan) })

	var addr string
	select {
	case addr = <-listening:
	case <-time.After(5 * time.Second):
		t.Fatal("HypeServer did not start listening")
	}

	body := make(chan string, 1)
	go func() {
		res, err := http.Get("http://" + addr)
		if !assert.NoError(t, err) {
			body <- ""
			return
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		body <- string(b)
	}()
	<-started

	stopped := make(chan error, 1)
	go func() { stopped <- hs.Stop(srvChan) }()
	select {
	case err := <-stopped:
		t.Fatalf("Stop returned while a request was in flight: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	close(release)
	require.NoError(t, <-stopped, "Stop should wait for the in-flight request")
	assert.Equal(t, "done", <-body, "the in-flight request completes")

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if conn != nil {
		conn.Close()
	}
	require.Error(t, err, "server should no longer accept connections after Stop")
}

func TestStats(t *testing.T) {
	store := testutils.NewStore(t)
	gs, us := guild.NewGuildStore(store), user.NewUserStore(store)
	hs, err := NewHypeServer(gs, us)
	require.NoError(t, err)

	checkStats := func(origin, wantJSON string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/stats", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		hs.server.Handler.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		assert.Equal(t, origin, res.Header.Get("Access-Control-Allow-Origin"))
		assert.JSONEq(t, wantJSON, rec.Body.String())
	}

	checkStats("", `{"servers":"0","users":"0"}`)

	gs.Add("guild-a")
	gs.Add("guild-b")
	gs.Remove("guild-b")
	us.Add(user.User{Guild_ID: "guild-a", UID: "user-a"})
	us.Add(user.User{Guild_ID: "guild-a", UID: "user-b"})
	us.Add(user.User{Guild_ID: "guild-b", UID: "user-c"})
	checkStats("https://dashboard.example", `{"servers":"2","users":"3"}`)

	// The server handles requests concurrently.
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			rec := httptest.NewRecorder()
			hs.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stats", nil))
			assert.JSONEq(t, `{"servers":"2","users":"3"}`, rec.Body.String())
		})
	}
	wg.Wait()
}
