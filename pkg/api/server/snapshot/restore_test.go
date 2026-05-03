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

func TestRestore_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/snapshot/restore.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("snapshot"); got != "91695" {
			t.Errorf("snapshot = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).Restore(context.Background(), SnapshotOptions{Name: "ch-foo", Snapshot: "91695"}); err != nil {
		t.Fatalf("Restore: %v", err)
	}
}

func TestRestore_RequiredFields(t *testing.T) {
	c, _ := api.New("k", "1")
	if _, err := New(c).Restore(context.Background(), SnapshotOptions{}); err == nil ||
		!strings.Contains(err.Error(), "Name is required") {
		t.Errorf("err = %v", err)
	}
}
