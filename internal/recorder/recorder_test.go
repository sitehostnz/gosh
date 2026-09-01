package recorder

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// do issues a request and closes the body, returning what came back.
// The linter requires a context on every call, and closing the body is
// the caller's job, so both live here rather than in every test.
func do(t *testing.T, client *http.Client, method, url, body string) string {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, url, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(out)
}

// doErr issues a request that is expected to fail at the transport.
func doErr(t *testing.T, client *http.Client, url string) error {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := client.Do(req)
	if err == nil {
		defer func() { _ = resp.Body.Close() }()
		t.Fatal("expected a transport error")
	}
	return err
}

// readRecordings loads every recording in dir, in filename order.
func readRecordings(t *testing.T, dir string) []Recording {
	t.Helper()
	names, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	out := make([]Recording, 0, len(names))
	for _, n := range names {
		raw, err := os.ReadFile(n) //nolint:gosec // test-controlled path
		if err != nil {
			t.Fatalf("read %s: %v", n, err)
		}
		var rec Recording
		if err := json.Unmarshal(raw, &rec); err != nil {
			t.Fatalf("decode %s: %v", n, err)
		}
		out = append(out, rec)
	}
	return out
}

// TestRecorder_CapturesTheAPIEnvelope is the point of the whole thing.
//
// This API answers HTTP 200 with {"status":false} when it rejects a
// request. A recording that kept only the status code would say the call
// succeeded — which is exactly the blind spot a hand-written mock has.
func TestRecorder_CapturesTheAPIEnvelope(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// A rejection, served with 200 as this API does.
		_, _ = io.WriteString(w, `{"status":false,"msg":"Error: Please specify a valid image type."}`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	do(t, client, http.MethodGet, srv.URL+"/1.5/server/list_images.json", "")

	recs := readRecordings(t, dir)
	if len(recs) != 1 {
		t.Fatalf("recordings = %d, want 1", len(recs))
	}
	rec := recs[0]
	if rec.Status != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Status)
	}
	if rec.OK {
		t.Error("api_ok = true for a rejection; the envelope is what distinguishes it from success")
	}
	if !strings.Contains(rec.Msg, "valid image type") {
		t.Errorf("api_msg = %q, want the rejection reason", rec.Msg)
	}
	if rec.When.IsZero() {
		t.Error("When is zero; a fixture with no date cannot be told from current truth")
	}
}

// TestRecorder_RedactsCredentials checks the apikey never reaches disk.
// It rides in the query string on every single call, so redaction has to
// be right every time rather than usually.
func TestRecorder_RedactsCredentials(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful"}`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	do(t, client, http.MethodPost, srv.URL+"/1.5/server/provision.json?apikey=supersecret&client_id=1",
		"password=hunter2&label=web")

	recs := readRecordings(t, dir)
	if len(recs) != 1 {
		t.Fatalf("recordings = %d, want 1", len(recs))
	}
	raw, err := json.Marshal(recs[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, secret := range []string{"supersecret", "hunter2"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("recording contains %q; credentials must be redacted at record time", secret)
		}
	}
	if !strings.Contains(recs[0].URL, "REDACTED") {
		t.Errorf("URL = %q, want the apikey replaced", recs[0].URL)
	}
	// The non-secret field must survive, or the recording is useless.
	var sawLabel bool
	for _, p := range recs[0].Form {
		if p.Key == "label" && p.Value == "web" {
			sawLabel = true
		}
	}
	if !sawLabel {
		t.Errorf("form = %+v, want label=web preserved", recs[0].Form)
	}
}

// TestRecorder_DoesNotConsumeTheBody checks the server still receives
// the request intact. A recorder that ate the body would break every
// write call it observed.
func TestRecorder_DoesNotConsumeTheBody(t *testing.T) {
	t.Parallel()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		_, _ = io.WriteString(w, `{"status":true}`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	do(t, client, http.MethodPost, srv.URL+"/1.5/server/provision.json", "label=web&location=AKLNCT")

	if got != "label=web&location=AKLNCT" {
		t.Errorf("server received %q, want the body intact", got)
	}
}

// TestRecorder_PreservesFormOrder matters because this API is
// order-sensitive in places, and url.ParseQuery would return a map.
func TestRecorder_PreservesFormOrder(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"status":true}`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	body := "client_id=1&label=web&location=AKLNCT&product_code=LHPVS1&image=x"
	do(t, client, http.MethodPost, srv.URL+"/x.json", body)

	recs := readRecordings(t, dir)
	want := []string{"client_id", "label", "location", "product_code", "image"}
	if len(recs[0].Form) != len(want) {
		t.Fatalf("form pairs = %d, want %d", len(recs[0].Form), len(want))
	}
	for i, k := range want {
		if recs[0].Form[i].Key != k {
			t.Errorf("form[%d] = %q, want %q — order is not preserved", i, recs[0].Form[i].Key, k)
		}
	}
}

// TestRecorder_SurvivesAnUnwritableDir checks the call still goes
// through when recording fails.
//
// A diagnostic that can break what it observes gets switched off, and
// then does not get used. The request must proceed regardless.
func TestRecorder_SurvivesAnUnwritableDir(t *testing.T) {
	t.Parallel()
	var reached bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful"}`)
	}))
	defer srv.Close()

	// A path under a regular file cannot be written to.
	blocker := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	client := &http.Client{Transport: New(filepath.Join(blocker, "nope"), nil)}
	body := do(t, client, http.MethodGet, srv.URL+"/x.json", "")

	if !reached {
		t.Error("the request never reached the server")
	}
	if !strings.Contains(body, "Successful") {
		t.Errorf("caller got %q, want the real response", body)
	}
}

// TestRecorder_RecordsTransportFailures checks a connection error is
// captured too — timeouts and resets are shapes a caller has to handle,
// and they are absent from every mock-based suite.
func TestRecorder_RecordsTransportFailures(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL + "/x.json"
	srv.Close() // nothing is listening now

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	_ = doErr(t, client, url)

	recs := readRecordings(t, dir)
	if len(recs) != 1 {
		t.Fatalf("recordings = %d, want 1 — transport failures are worth recording", len(recs))
	}
	if recs[0].Status != 0 {
		t.Errorf("Status = %d, want 0 for a call that never got a response", recs[0].Status)
	}
}

// TestRecorder_RedactsNestedAndResponseSecrets covers the two gaps the
// original redaction had.
//
// The request side matched field names exactly against a three-entry
// list, so it caught "password" and missed "params[password]" — which
// is the only spelling this API actually uses. The test that was meant
// to cover it sent the one spelling on the list, so it asserted the
// belief that produced the list rather than what goes on the wire.
//
// The response side was not redacted at all, and server.Create returns
// the new machine's root password, which the API emits once and never
// again.
func TestRecorder_RedactsNestedAndResponseSecrets(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// The shape server/provision.json actually returns.
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"name":"gosh-journey","password":"R00tS3cret!","job":{"id":1}}}`)
	}))
	defer srv.Close()

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	do(t, client, http.MethodPost, srv.URL+"/1.5/server/provision.json?apikey=REALKEY&client_id=1",
		"params%5Bpassword%5D=Sup3rS3cret%21&label=web")

	recs := readRecordings(t, dir)
	if len(recs) != 1 {
		t.Fatalf("recordings = %d, want 1", len(recs))
	}
	raw, err := json.Marshal(recs[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	for _, secret := range []string{"Sup3rS3cret!", "REALKEY", "R00tS3cret!"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("recording contains %q; every secret must be redacted at record time", secret)
		}
	}

	// The non-secret values must survive, or the recording is useless
	// as a fixture.
	if !strings.Contains(recs[0].Body, "gosh-journey") {
		t.Errorf("body = %q, want the non-secret fields preserved", recs[0].Body)
	}
	var sawLabel bool
	for _, p := range recs[0].Form {
		if p.Key == "label" && p.Value == "web" {
			sawLabel = true
		}
	}
	if !sawLabel {
		t.Errorf("form = %+v, want label=web preserved", recs[0].Form)
	}
}

// TestRecorder_KeepsANonJSONBody checks an unparseable response is
// still recorded. Dropping it would lose exactly the case worth having.
func TestRecorder_KeepsANonJSONBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "<html>502 Bad Gateway</html>")
	}))
	defer srv.Close()

	dir := t.TempDir()
	client := &http.Client{Transport: New(dir, nil)}
	do(t, client, http.MethodGet, srv.URL+"/x.json", "")

	recs := readRecordings(t, dir)
	if len(recs) != 1 || !strings.Contains(recs[0].Body, "502 Bad Gateway") {
		t.Errorf("body = %q, want the unparseable response kept verbatim", recs[0].Body)
	}
}
