package hypeserver

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sonastea/hypebot/internal/database"
	"github.com/sonastea/hypebot/internal/datastore/guild"
	"github.com/sonastea/hypebot/internal/datastore/user"
	"github.com/sonastea/hypebot/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var (
	err error

	db        *sql.DB
	hs        *HypeServer
	mockStore *testutils.MockStore
)

func TestMain(m *testing.M) {
	db, _ = sql.Open("sqlite3", ":memory:")

	mockStore = testutils.NewMockStore(db)

	_, err := db.Exec(database.Schema)
	if err != nil {
		log.Println(err)
	}

	os.Exit(m.Run())
}

func TestNewHypeServer(t *testing.T) {
	hs, err = NewHypeServer(guild.NewGuildStore(mockStore), user.NewUserStore(mockStore))
	if err != nil {
		t.Fatalf("unable to create hypeserver: %s", err)
	}

	assert.IsType(t, &HypeServer{}, hs)
	assert.Equal(t, uint64(0), hs.servers)
	assert.Equal(t, uint64(0), hs.users)
}

func TestRunAndStop(t *testing.T) {
	ctx, srvChan := hs.Run()
	assert.NotNil(t, srvChan, "unable to return os.Signal from running hypeserver")

	err := hs.Stop(ctx, srvChan)
	assert.Nil(t, err, "unable to shut down hypeserver gracefully: %v", err)
}

func TestStats(t *testing.T) {
	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatalf("get request error for /stats: %v", err)
	}

	w := httptest.NewRecorder()
	stats(guild.NewGuildStore(mockStore), user.NewUserStore(mockStore)).ServeHTTP(w, req)

	assert.Exactlyf(t, w.Code, http.StatusOK, "stats handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)

	expected := `{"servers":"0","users":"0"}
`
	assert.Exactlyf(t, expected, w.Body.String(), "/stats returned unexecpted body: got %v want %v", w.Body.String(), expected)
}

func TestEnableCors(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("get request error for /: %v", err)
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w, r)
	})
	handler.ServeHTTP(w, req)

	assert.Exactlyf(t, "", w.Header().Get("Access-Control-Allow-Origin"), "enable cors return unexpected header: got %v want %v", w.Header().Get("Access-Control-Allow-Origin"), "")
}
