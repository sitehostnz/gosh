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

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("partition"); got != "scsi0" {
			t.Errorf("partition = %q", got)
		}
		if got := r.Form.Get("lifetime"); got != "24" {
			t.Errorf("lifetime = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":29658457,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Create(context.Background(), CreateOptions{
		Name: "ch-foo", Partition: "scsi0", Lifetime: 24,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Return.ID != 29658457 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestCreate_RequiredFields(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).Create(context.Background(), CreateOptions{Partition: "scsi0", Lifetime: 1}); err == nil ||
		!strings.Contains(err.Error(), "Name is required") {
		t.Errorf("missing Name: err = %v", err)
	}
	if _, err := New(c).Create(context.Background(), CreateOptions{Name: "n", Lifetime: 1}); err == nil ||
		!strings.Contains(err.Error(), "Partition is required") {
		t.Errorf("missing Partition: err = %v", err)
	}
	if _, err := New(c).Create(context.Background(), CreateOptions{Name: "n", Partition: "scsi0"}); err == nil ||
		!strings.Contains(err.Error(), "Lifetime must be > 0") {
		t.Errorf("missing Lifetime: err = %v", err)
	}
}
