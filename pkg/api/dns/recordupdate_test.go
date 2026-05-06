package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestUpdateRecord_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/dns/update_record.json" {
			t.Errorf("path = %q, want /dns/update_record.json", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("record_id"); got != "9000001" {
			t.Errorf("record_id = %q, want 9000001", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status": true, "msg": "Successful"}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).UpdateRecord(context.Background(), UpdateRecordRequest{
		Domain:   testDomain,
		RecordID: "9000001",
		Type:     "A",
		Name:     "www.example.co.nz",
		Content:  "192.0.2.20",
	})
	if err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}

	if !got.Status {
		t.Errorf("Status = false, want true")
	}
}
