package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/job"
	"github.com/sitehostnz/gosh/pkg/api/server"
)

// fakeAPI serves just enough of the API for the rollback path: an
// AddIP that queues a job, and a job that is already finished.
//
// addIPFails makes AddIP refuse, which is the case worth testing —
// a restore attempted while the API is unhappy is the likely one, since
// whatever broke the swap has probably not gone away.
func fakeAPI(t *testing.T, addIPFails bool) clients {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "add_ip"):
			if addIPFails {
				// This API reports refusals as HTTP 200 with
				// status:false, so the fake has to as well.
				_, _ = w.Write([]byte(`{"status":false,"msg":"The ip address is invalid, please specify a valid ip address"}`))
				return
			}
			// The job id is a bare number, not a quoted one. Confirmed
			// by the journey running green against the live API — a
			// quoted id would have failed waitJob's decode. Writing it
			// quoted here produced a test failure that looked like a
			// bug in restoreHeld, which is the hand-written-fixture
			// trap in miniature.
			_, _ = w.Write([]byte(`{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"}}}`))
		case strings.Contains(r.URL.Path, "job/get"):
			_, _ = w.Write([]byte(`{"status":true,"msg":"Successful","return":{"state":"Completed"}}`))
		default:
			t.Errorf("unexpected call to %s", r.URL.Path)
			_, _ = w.Write([]byte(`{"status":false,"msg":"unexpected"}`))
		}
	}))
	t.Cleanup(srv.Close)

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	return clients{server: server.New(c), job: job.New(c)}
}

// captureLog collects what a function logs.
func captureLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)
	fn()
	return buf.String()
}

// TestRestoreHeld_PutsAReleasedAddressBack covers the path that only
// runs when a swap fails part-way.
//
// Between releasing an address and assigning its replacement a server
// holds nothing and is reachable only by console. Inside the journey
// that is contained, because cleanup deletes both servers; standalone
// against a pair someone already owns it is a production outage. The
// happy path is exercised by running the journey, but this path never
// is — which is exactly why it needs a test.
//
//nolint:paralleltest // captureLog swaps the global log writer; running these in parallel would interleave their output
func TestRestoreHeld_PutsAReleasedAddressBack(t *testing.T) {
	c := fakeAPI(t, false)
	held := map[string]string{"server-a": "203.0.113.10"}

	out := captureLog(t, func() {
		restoreHeld(context.Background(), c, held)
	})

	if !strings.Contains(out, "restoring 203.0.113.10") {
		t.Errorf("log does not announce the restore:\n%s", out)
	}
	if !strings.Contains(out, "✓ restored 203.0.113.10 to server-a") {
		t.Errorf("log does not confirm the restore succeeded:\n%s", out)
	}
}

// TestRestoreHeld_SaysWhatToDoByHandWhenItCannotRestore is the case
// that matters most. If the restore itself fails, the operator has a
// server holding no address, and the only useful thing this can do is
// name the exact call that fixes it.
//
//nolint:paralleltest // captureLog swaps the global log writer; running these in parallel would interleave their output
func TestRestoreHeld_SaysWhatToDoByHandWhenItCannotRestore(t *testing.T) {
	c := fakeAPI(t, true)
	held := map[string]string{"server-a": "203.0.113.10"}

	out := captureLog(t, func() {
		restoreHeld(context.Background(), c, held)
	})

	if !strings.Contains(out, "could not restore") {
		t.Errorf("a failed restore must say so:\n%s", out)
	}
	if !strings.Contains(out, "server.AddIP(server-a, 203.0.113.10)") {
		t.Errorf("a failed restore must name the call that fixes it by hand:\n%s", out)
	}
	if strings.Contains(out, "✓ restored") {
		t.Errorf("a failed restore must not report success:\n%s", out)
	}
}

// TestRestoreHeld_DoesNothingWhenNothingIsHeld guards the ordinary
// case: a swap that completed owes nothing, and the deferred call must
// stay silent rather than logging an alarming line on every run.
//
//nolint:paralleltest // captureLog swaps the global log writer; running these in parallel would interleave their output
func TestRestoreHeld_DoesNothingWhenNothingIsHeld(t *testing.T) {
	c := fakeAPI(t, false)

	out := captureLog(t, func() {
		restoreHeld(context.Background(), c, map[string]string{})
	})

	if out != "" {
		t.Errorf("a completed swap must restore nothing and say nothing, got:\n%s", out)
	}
}
