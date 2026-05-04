package volume

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGet_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/volume/get.json" {
			t.Errorf("path = %q, want /cloud/volume/get.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server"); got != testServerName {
			t.Errorf("server = %q, want ch-foo (note: not server_name)", got)
		}
		if got := r.URL.Query().Get("volume"); got != testVolumeName {
			t.Errorf("volume = %q, want data-vol (note: not volume_name)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful",
			"return": {
				"id": "100", "client_id": "1234", "server_id": "62994",
				"pending": "", "volume_name": "data-vol",
				"is_missing": "0",
				"date_added": "2026-05-01 00:00:00", "date_updated": "2026-05-02 00:00:00",
				"server_name": "ch-foo", "server_label": "foo",
				"server_owner": true, "containers": []
			}
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Get(context.Background(), GetOptions{Server: testServerName, Volume: testVolumeName})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Return.VolumeName != testVolumeName {
		t.Errorf("VolumeName = %q", got.Return.VolumeName)
	}
}

func TestGet_RequiredFields(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).Get(context.Background(), GetOptions{Volume: "v"}); err == nil ||
		!strings.Contains(err.Error(), "Server is required") {
		t.Errorf("missing Server: err = %v", err)
	}
	if _, err := New(c).Get(context.Background(), GetOptions{Server: "s"}); err == nil ||
		!strings.Contains(err.Error(), "Volume is required") {
		t.Errorf("missing Volume: err = %v", err)
	}
}
