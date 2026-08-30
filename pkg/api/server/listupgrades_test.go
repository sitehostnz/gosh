package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// serveFixture serves a recorded response for any path.
func serveFixture(t *testing.T, name string) *api.Client {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	return c
}

// TestListUpgrades_DecodesARecordedResponse is built on a real response
// rather than a hand-written one, because a hand-written one is what
// hid this bug.
//
// The endpoint had two type errors. The first — int quota fields where
// the API sends 67.5 — was found and fixed. The call was then never
// re-run, so the second went unnoticed and the endpoint was documented
// as corrected while still failing outright on:
//
//	cannot unmarshal string into Go struct field
//	Upgrades.return.cores of type int
//
// Note the asymmetry that makes it hard to guess: cores arrives quoted
// and ram arrives bare, in the same object.
func TestListUpgrades_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	c := serveFixture(t, "list_upgrades.json")

	got, err := New(c).ListUpgrades(context.Background(), ListUpgradesOptions{Name: "s"})
	if err != nil {
		t.Fatalf("ListUpgrades: %v", err)
	}

	if len(got.Return.Cores) == 0 {
		t.Fatal("Cores is empty; the API sends quoted integers here")
	}
	if got.Return.Cores[0] != 8 {
		t.Errorf(`Cores[0] = %d, want 8 decoded from the string "8"`, got.Return.Cores[0])
	}
	if len(got.Return.RAM) == 0 || got.Return.RAM[0] != 4 {
		t.Errorf("RAM = %v, want [4] — sent as bare numbers alongside quoted cores", got.Return.RAM)
	}
	if len(got.Return.Plan) == 0 {
		t.Fatal("Plan is empty; it is the authoritative list of what this server may become")
	}
	if got.Return.Plan[0] != "LHPVS2" {
		t.Errorf("Plan[0] = %q, want LHPVS2", got.Return.Plan[0])
	}

	// The quota fields are the half that was already fixed; keeping the
	// assertion stops a later tidy-up putting them back to int.
	if got.Return.Quota.RAM.Used != 65.5 {
		t.Errorf("Quota.RAM.Used = %v, want 65.5 — fractional, so not an int", got.Return.Quota.RAM.Used)
	}

	// Used above Total is a real state, not a decode error: it is what
	// an over-quota account looks like.
	if got.Return.Quota.Cores.Used <= got.Return.Quota.Cores.Total {
		t.Log("note: this fixture no longer exercises the over-quota case")
	}

	// The disk labels are keyed per disk, which is the shape
	// UpgradeComponents needs.
	if _, ok := got.Return.Disk["scsi0"]; !ok {
		t.Errorf("Disk has no scsi0 entry; keys = %v", keysOf(got.Return.Disk))
	}
}

// keysOf lists a map's keys for an error message.
func keysOf(m map[string]DiskUpgradeOptions) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
