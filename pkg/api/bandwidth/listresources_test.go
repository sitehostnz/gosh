package bandwidth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListResources_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bandwidth/list_resources.json" {
			t.Errorf("path = %q, want /bandwidth/list_resources.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{
					"client_id": "1234",
					"group_id": "1",
					"group_name": "NZ - Linux Servers",
					"quotas": [
						{
							"attribute_id": "1",
							"attribute_name": "VPS Disk Space",
							"attribute_unit": "GB",
							"attribute_type": "0",
							"total_units": "1424",
							"used_units": "1524",
							"available_units": -100,
							"objects": ["server-a", "server-b"]
						}
					]
				}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListResources(context.Background())
	if err != nil {
		t.Fatalf("ListResources: %v", err)
	}

	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	g := got.Return[0]
	if g.GroupName != "NZ - Linux Servers" {
		t.Errorf("GroupName = %q, want NZ - Linux Servers", g.GroupName)
	}
	if len(g.Quotas) != 1 {
		t.Fatalf("len(Quotas) = %d, want 1", len(g.Quotas))
	}
	if g.Quotas[0].AvailableUnits != -100 {
		t.Errorf("AvailableUnits = %d, want -100", g.Quotas[0].AvailableUnits)
	}
}
