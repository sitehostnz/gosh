package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestDeleteZone_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/dns/delete_domain.json" {
			t.Errorf("path = %q, want /dns/delete_domain.json", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("domain"); got != "example.co.nz" {
			t.Errorf("domain = %q, want example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status": true, "msg": "Successful"}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).DeleteZone(context.Background(), DeleteZoneRequest{DomainName: "example.co.nz"})
	if err != nil {
		t.Fatalf("DeleteZone: %v", err)
	}

	if !got.Status {
		t.Errorf("Status = false, want true")
	}
}
