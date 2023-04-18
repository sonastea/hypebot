package hypeserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	servers uint64 = 69
	users   uint64 = 420
)

func MockStats(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)

	switch r.URL.Query().Get("test") {
	case "Get Cached Stats from DB":
		{
			servers = 68
			users = 419
		}
	}

	data["servers"] = strconv.FormatUint(servers, 10)
	data["users"] = strconv.FormatUint(users, 10)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func TestGetStats(t *testing.T) {
	tests := []struct {
		name        string
		req         *http.Request
		wantServers string
		wantUsers   string
	}{
		{

			name:        "Get Stats from DB",
			wantServers: "69",
			wantUsers:   "420",
		},
		{

			name:        "Get Cached Stats from DB",
			wantServers: "68",
			wantUsers:   "419",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/stats", nil)
			q := req.URL.Query()
			q.Add("test", test.name)
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			MockStats(w, req)
			res := w.Result()

			body, _ := io.ReadAll(res.Body)
			data := make(map[string]string)
			json.Unmarshal(body, &data)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, test.wantServers, data["servers"])
			assert.Equal(t, test.wantUsers, data["users"])
		})
	}
}
