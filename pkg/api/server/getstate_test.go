package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGetState_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/server/get_state.json" {
			t.Errorf("path = %q, want /server/get_state.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "ch-foo" {
			t.Errorf("name = %q, want ch-foo", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"state": "On",
				"rescue": false,
				"last_job": {"id": "29426222", "type": "scheduler", "state": "Complete"}
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetState(context.Background(), GetStateOptions{Name: "ch-foo"})
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}

	if got.Return.State != "On" {
		t.Errorf("State = %q, want On", got.Return.State)
	}
	if got.Return.Rescue {
		t.Errorf("Rescue = true, want false")
	}
	if got.Return.LastJob.ID != "29426222" {
		t.Errorf("LastJob.ID = %q, want 29426222", got.Return.LastJob.ID)
	}
	if got.Return.LastJob.State != "Complete" {
		t.Errorf("LastJob.State = %q, want Complete", got.Return.LastJob.State)
	}
}
