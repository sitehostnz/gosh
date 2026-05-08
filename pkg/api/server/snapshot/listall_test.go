package snapshot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestList_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/snapshot/list_all.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "ch-foo" {
			t.Errorf("name = %q (note: not server_name)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful",
			"return": [
				{
					"id": "91695", "name": "s-20260502135000",
					"device": "s-20260502135000", "mountpoint": "",
					"size": 0, "fstype": "raw", "drbd": "0", "parent": "90875",
					"pending": "0", "created": "2026-05-02 13:50:00",
					"backup": "0",
					"disk_total": "0", "disk_used": "0",
					"inodes_total": "0", "inodes_used": "0",
					"stats_updated": "0000-00-00 00:00:00",
					"disk_warn": "0", "is_missing": false,
					"expires": "2026-05-02 14:50:00"
				}
			]
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).List(context.Background(), ListOptions{Name: "ch-foo"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	s := got.Return[0]
	if s.ID != "91695" {
		t.Errorf("ID = %q", s.ID)
	}
	if s.IsMissing {
		t.Errorf("IsMissing = true, want false (real bool)")
	}
	if s.Pending != "0" {
		t.Errorf("Pending = %q (string-typed)", s.Pending)
	}
}

func TestList_NameRequired(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).List(context.Background(), ListOptions{}); err == nil ||
		!strings.Contains(err.Error(), "Name is required") {
		t.Errorf("err = %v", err)
	}
}
