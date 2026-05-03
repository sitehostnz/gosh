package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestAddRecord_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/dns/add_record.json" {
			t.Errorf("path = %q, want /dns/add_record.json", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("domain"); got != "example.co.nz" {
			t.Errorf("domain = %q, want example.co.nz", got)
		}
		if got := r.Form.Get("type"); got != "A" {
			t.Errorf("type = %q, want A", got)
		}
		if got := r.Form.Get("content"); got != "192.0.2.10" {
			t.Errorf("content = %q, want 192.0.2.10", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status": true, "msg": "Successful", "return": {"id": "9000001"}}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).AddRecord(context.Background(), AddRecordRequest{
		Domain:  "example.co.nz",
		Type:    "A",
		Name:    "www.example.co.nz",
		Content: "192.0.2.10",
	})
	if err != nil {
		t.Fatalf("AddRecord: %v", err)
	}

	if got.Return.ID != "9000001" {
		t.Errorf("Return.ID = %q, want 9000001", got.Return.ID)
	}
}
