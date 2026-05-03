package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGetZone_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/dns/search_domains.json" {
			t.Errorf("path = %q, want /dns/search_domains.json", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("query[domain]"); got != "example.co.nz" {
			t.Errorf("query[domain] = %q, want example.co.nz", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{"client_id": "1234", "name": "example.co.nz", "template_id": "0"}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetZone(context.Background(), GetZoneRequest{DomainName: "example.co.nz"})
	if err != nil {
		t.Fatalf("GetZone: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	z := got.Return[0]
	if z.Name != "example.co.nz" {
		t.Errorf("Return[0].Name = %q, want example.co.nz", z.Name)
	}
	if z.ClientID != "1234" {
		t.Errorf("Return[0].ClientID = %q, want 1234", z.ClientID)
	}
	if z.TemplateID != "0" {
		t.Errorf("Return[0].TemplateID = %q, want 0", z.TemplateID)
	}
}
