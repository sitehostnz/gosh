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
	t.Parallel()
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

// TestListResources_MixedNumberShapes locks down the mixed-type quirk:
// the API returns used_units as a JSON string when usage is non-zero
// and as a JSON number when usage is zero, within a single response.
// Both forms must decode without error.
func TestListResources_MixedNumberShapes(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
							"attribute_name": "VPS Disk Space (used)",
							"attribute_unit": "GB",
							"attribute_type": "0",
							"total_units": "1424",
							"used_units": "783",
							"available_units": 641,
							"objects": []
						},
						{
							"attribute_id": "2",
							"attribute_name": "VPS Disk Space (zero)",
							"attribute_unit": "GB",
							"attribute_type": "0",
							"total_units": "1424",
							"used_units": 0,
							"available_units": 1424,
							"objects": []
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
	if len(got.Return) != 1 || len(got.Return[0].Quotas) != 2 {
		t.Fatalf("unexpected shape: %+v", got.Return)
	}
	if got.Return[0].Quotas[0].UsedUnits != 783 {
		t.Errorf("Quotas[0].UsedUnits = %v, want 783", got.Return[0].Quotas[0].UsedUnits)
	}
	if got.Return[0].Quotas[1].UsedUnits != 0 {
		t.Errorf("Quotas[1].UsedUnits = %v, want 0", got.Return[0].Quotas[1].UsedUnits)
	}
}
