package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListUpgrades_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/list_upgrades.json" {
			t.Errorf("path = %q, want /server/list_upgrades.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("name"); got != "ch-foo" {
			t.Errorf("name = %q, want ch-foo", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"quota": {
					"ram":   {"total": 38, "used": 38},
					"disk":  {"total": 1424, "used": 1524},
					"cores": {"total": 27, "used": 18}
				},
				"extra-disk": {"price": 2.5, "size": 5},
				"disk": {
					"scsi0": {"included": [50], "extra": [55, 60, 65]}
				}
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListUpgrades(context.Background(), ListUpgradesOptions{Name: "ch-foo"})
	if err != nil {
		t.Fatalf("ListUpgrades: %v", err)
	}

	if got.Return.Quota.RAM.Total != 38 {
		t.Errorf("Quota.RAM.Total = %d, want 38", got.Return.Quota.RAM.Total)
	}
	if got.Return.Quota.Cores.Used != 18 {
		t.Errorf("Quota.Cores.Used = %d, want 18", got.Return.Quota.Cores.Used)
	}
	if got.Return.ExtraDisk.Price != 2.5 {
		t.Errorf("ExtraDisk.Price = %v, want 2.5", got.Return.ExtraDisk.Price)
	}
	scsi0, ok := got.Return.Disk["scsi0"]
	if !ok {
		t.Fatal("Disk[scsi0] missing")
	}
	if len(scsi0.Extra) != 3 {
		t.Errorf("Disk[scsi0].Extra len = %d, want 3", len(scsi0.Extra))
	}
}
