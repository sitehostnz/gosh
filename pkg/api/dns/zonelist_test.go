package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListZones_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/list_domains.json" {
			t.Errorf("path = %q, want /dns/list_domains.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"current_items": 1,
				"current_page": 1,
				"total_pages": 1,
				"total_items": 1,
				"data": [
					{"name": "example.co.nz", "client_id": "1234", "template_id": "0", "pending": "0"}
				]
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListZones(context.Background(), &ListZoneOptions{})
	if err != nil {
		t.Fatalf("ListZones: %v", err)
	}

	if len(got.Return.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(got.Return.Data))
	}
	z := got.Return.Data[0]
	if z.Name != testDomain {
		t.Errorf("Name = %q, want %s", z.Name, testDomain)
	}
	if z.TemplateID != "0" {
		t.Errorf("TemplateID = %q, want 0", z.TemplateID)
	}
}
