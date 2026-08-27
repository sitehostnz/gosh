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
					"ram":   {"total": 67.5, "used": 67.5},
					"disk":  {"total": 1424, "used": 1524},
					"cores": {"total": 27, "used": 18}
				},
				"extra-disk": {"price": 2.5, "size": 5},
				"disk": {
					"scsi0": {"included": [50], "extra": [55, 60, 65]}
				},
				"cores": [1],
				"ram": [2],
				"plan": ["LHPVS1", "LHPVS2", "LHPVS4"]
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

	// The RAM quota is fractional on real responses; this fixture said
	// 38 and declaring the field as int meant the endpoint never decoded
	// at all against a live account.
	if got.Return.Quota.RAM.Total != 67.5 {
		t.Errorf("Quota.RAM.Total = %v, want 67.5", got.Return.Quota.RAM.Total)
	}
	if got.Return.Quota.Cores.Used != 18 {
		t.Errorf("Quota.Cores.Used = %v, want 18", got.Return.Quota.Cores.Used)
	}
	// Used exceeding Total is a legitimate over-quota state.
	if got.Return.Quota.Disk.Used <= got.Return.Quota.Disk.Total {
		t.Error("expected the fixture's over-quota disk to survive decoding")
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

	// Only resizable disks appear; a swap partition is absent rather
	// than present with empty options.
	if _, present := got.Return.Disk["scsi1"]; present {
		t.Error("Disk[scsi1] should be absent, not present-and-empty")
	}

	// The three fields the endpoint returns that were previously
	// discarded. These are what UpgradeComponents and Upgrade validate
	// against, so a caller cannot choose a legal value without them.
	if len(got.Return.Cores) != 1 || got.Return.Cores[0] != 1 {
		t.Errorf("Cores = %v, want [1]", got.Return.Cores)
	}
	if len(got.Return.RAM) != 1 || got.Return.RAM[0] != 2 {
		t.Errorf("RAM = %v, want [2]", got.Return.RAM)
	}
	if len(got.Return.Plan) != 3 || got.Return.Plan[0] != "LHPVS1" {
		t.Errorf("Plan = %v, want [LHPVS1 LHPVS2 LHPVS4]", got.Return.Plan)
	}
}
