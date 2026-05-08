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

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("snapshot"); got != "91695" {
			t.Errorf("snapshot = %q (note: not snapshot_id)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).Delete(context.Background(), SnapshotOptions{Name: "ch-foo", Snapshot: "91695"}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_RequiredFields(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).Delete(context.Background(), SnapshotOptions{Snapshot: "1"}); err == nil ||
		!strings.Contains(err.Error(), "Name is required") {
		t.Errorf("missing Name: err = %v", err)
	}
	if _, err := New(c).Delete(context.Background(), SnapshotOptions{Name: "n"}); err == nil ||
		!strings.Contains(err.Error(), "Snapshot is required") {
		t.Errorf("missing Snapshot: err = %v", err)
	}
}
