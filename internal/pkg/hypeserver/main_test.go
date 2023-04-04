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

func (s *HypeServer) mockStats(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)

	switch r.URL.Query().Get("test") {
	case "2":
		{
			s.servers = 68
			s.users = 419
		}
	}

	data["servers"] = strconv.FormatUint(s.servers, 10)
	data["users"] = strconv.FormatUint(s.users, 10)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func TestGetStats(t *testing.T) {
	hs := &HypeServer{
		server: &http.Server{
			Addr: ":3000",
		},
		servers: 69,
		users:   420,
	}

	tests := []struct {
		name        string
		s           *HypeServer
		req         *http.Request
		wantServers string
		wantUsers   string
	}{
		{

			name: "Get Stats from DB",
			s:    hs,
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/stats", nil)
				q := req.URL.Query()
				q.Add("test", "1")
				req.URL.RawQuery = q.Encode()
				return req
			}(),
			wantServers: "69",
			wantUsers:   "420",
		},
		{

			name: "Get Cached Stats from DB",
			s:    hs,
			req: func() *http.Request {
				req := httptest.NewRequest("GET", "/stats", nil)
				q := req.URL.Query()
				q.Add("test", "2")
				req.URL.RawQuery = q.Encode()
				return req

			}(),
			wantServers: "68",
			wantUsers:   "419",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			test.s.mockStats(w, test.req)

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
