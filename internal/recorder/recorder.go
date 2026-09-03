// Package recorder captures real API exchanges so that test fixtures can
// be derived from evidence rather than from belief.
//
// It lives in internal/ deliberately: it is a development tool, and an
// exported option on a public SDK would be permanent surface that
// consumers should never call. Install it with api.SetTransport.
package recorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sitehostnz/gosh/pkg/models"
)

// Recording is one request and what the API said back.
//
// # Why this exists
//
// The unit tests here answer requests from hand-written mocks, and a
// mock accepts whatever it is given. That makes a wrong parameter name
// or form structurally invisible: the code and the mock agree, the test
// is green, and the call fails against the real API.
//
// Two live examples, both found in one afternoon:
//
//   - Create sent params[ipv4][]=auto for automatic allocation. The API
//     wants the scalar params[ipv4]. Every provision that let the SDK
//     choose an address failed, while TestCreate_DefaultsToAutoIPv4
//     passed — because it asserted what we send rather than what is
//     accepted.
//   - UpgradeComponents declared the disk result as a bool where the
//     API answers per disk label. Every disk upgrade failed to decode,
//     and the fixture asserting "disk": true passed throughout — it had
//     been written from the same belief as the struct.
//
// A third case shows the other half of the problem. A session reported
// that filters[type]=hpvm-distro was rejected at every location and
// concluded the HPVS discovery path was dead. It was not: the scalar
// form works and the array form does not, and no recording existed to
// settle it either way. Recordings are as useful for disproving a
// reported breakage as for finding one.
//
// # What a recording is for
//
// Fixtures built from recordings are evidence. Fixtures written by hand
// restate the belief that produced the bug.
//
// The valuable half is the REJECTIONS. A recorded success only tells
// you what came back for a request that already worked; neither bug
// above would appear in a corpus of successes. Errors are also the
// cheapest thing to record — nothing is provisioned and nothing has a
// side effect — so they can be refreshed often.
//
// # What it does not do
//
// It records what the SDK sends. If the SDK sends the wrong shape, the
// recording faithfully preserves the wrong shape. Discovering that a
// DIFFERENT form is the accepted one still takes somebody trying it.
// This catches drift and decode surprises; it does not replace probing.
type Recording struct {
	// When the call was made, so a stale fixture is visible as stale
	// rather than being taken for current truth.
	When time.Time `json:"when"`

	Method   string `json:"method"`
	Endpoint string `json:"endpoint"`
	// URL is redacted: the API key travels as a query parameter.
	URL string `json:"url"`
	// Form is the request body, parsed and redacted. Recorded as an
	// ordered list rather than a map because this API is
	// order-sensitive in places, and a map would lose that.
	Form []FormPair `json:"form,omitempty"`

	Status int    `json:"status"`
	Body   string `json:"body"`
	// OK reflects the API's own status field, which is not the HTTP
	// status: this API answers 200 with {"status":false} on rejection.
	// Recording both is the point — an error case with HTTP 200 is
	// exactly the shape a hand-written mock forgets.
	OK  bool   `json:"api_ok"`
	Msg string `json:"api_msg,omitempty"`
}

// FormPair is one submitted field.
type FormPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// redactTokens are the substrings that mark a field as secret.
//
// Matched by containment rather than equality, and that is the whole
// point. An exact-match denylist held "password" and missed
// params[password], which is the spelling this API actually uses for
// every password it accepts — cloud/ssh/user.Add, .Update and the mail
// endpoints all send it nested. So the one entry that mattered was the
// one that never matched, and a recording carried the value in
// plaintext.
//
// The test that was supposed to cover this sent password=hunter2: the
// single spelling on the denylist, and the one spelling the SDK never
// sends. It asserted the belief that produced the list rather than what
// goes on the wire, which is the failure this whole package exists to
// defeat, committed by the package itself.
var redactTokens = []string{"apikey", "api_key", "password", "passwd", "secret", "token"}

// shouldRedact reports whether a field name marks a secret.
func shouldRedact(key string) bool {
	k := strings.ToLower(key)
	for _, token := range redactTokens {
		if strings.Contains(k, token) {
			return true
		}
	}
	return false
}

// New returns a RoundTripper that records every exchange into dir, one
// JSON file per call, wrapping next. A nil next uses
// http.DefaultTransport.
//
// # A recording is not safe to share
//
// Secret-bearing fields are blanked on the way to disk, on both the
// request and the response, matched by key rather than by value. That
// is a reduction in exposure and not a guarantee: a recording still
// holds live customer data — server names and labels, addresses,
// database names, usernames, home directories, key material — and a
// field naming a secret in a way these tokens do not anticipate will
// go to disk in the clear.
//
// So: keep recordings off the repository and off anything shared, and
// pass them through [Scrub] before committing or attaching anything
// derived from them. Point dir outside the working tree; the examples
// use a temporary directory for that reason.
//
// Intended for use while implementing against the live API — which is
// when the interesting rejections happen anyway — so that fixtures can
// be derived from what the API actually did rather than composed from
// what we expected.
//
//	c, err := api.New(key, id,
//		api.SetTransport(recorder.New("/tmp/gosh-recordings", nil)))
//
// Recording never fails the call it observes: write errors go to stderr
// and the request proceeds. A diagnostic that can break what it watches
// gets switched off, and then does not get used.
func New(dir string, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	// Create the directory here rather than leaving every caller to do
	// it. Without this, copying the example above produces no
	// recordings and one stderr line per call, which is a confusing way
	// to discover a missing directory. Reported rather than returned:
	// this is a diagnostic, and it must not fail construction.
	if err := os.MkdirAll(dir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "gosh: recorder: %v\n", err)
	}
	return &recordingTransport{dir: dir, base: next}
}

type recordingTransport struct {
	base http.RoundTripper
	dir  string
	seq  atomic.Uint32
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := Recording{
		When:     time.Now().UTC(),
		Method:   req.Method,
		Endpoint: strings.TrimPrefix(req.URL.Path, "/"),
		URL:      models.RedactURL(req.URL),
	}

	// Read the body so it can be recorded, then put it back: a
	// RoundTripper that consumes the body breaks the request it is
	// observing.
	if req.Body != nil {
		raw, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			// Do not fail the call. The invariant this package states —
			// that recording never breaks what it observes — has to
			// hold without a reader checking, and returning here was
			// the one path that broke it. Record what was read and let
			// the request proceed with it.
			fmt.Fprintf(os.Stderr, "gosh: recorder: reading request body: %v\n", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(raw))
		rec.Form = parseForm(string(raw))
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		rec.Status = 0
		rec.Body = err.Error()
		t.write(rec)
		return resp, err
	}

	raw, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(raw))
	if readErr == nil {
		rec.Status = resp.StatusCode
		rec.Body = redactBody(raw)
		var envelope struct {
			Status bool   `json:"status"`
			Msg    string `json:"msg"`
		}
		if json.Unmarshal(raw, &envelope) == nil {
			rec.OK, rec.Msg = envelope.Status, envelope.Msg
		}
	}
	t.write(rec)
	return resp, nil
}

// write names the file so a directory listing reads as a transcript:
// endpoint, whether it was accepted, and a counter to keep order.
func (t *recordingTransport) write(rec Recording) {
	// Three outcomes, not two. A connection that never reached the API
	// is neither accepted nor rejected, and a directory listing should
	// say which — the same distinction the probe step goes to trouble
	// to make.
	outcome := "ok"
	if !rec.OK {
		outcome = "rejected"
	}
	if rec.Status == 0 {
		outcome = "notreached"
	}
	name := fmt.Sprintf("%03d-%s-%s.json",
		t.seq.Add(1),
		strings.NewReplacer("/", "_", ".json", "").Replace(rec.Endpoint),
		outcome)

	body, err := json.MarshalIndent(rec, "", "  ")
	if err == nil {
		err = os.WriteFile(filepath.Join(t.dir, name), append(body, '\n'), 0o600)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gosh: recorder: %v\n", err)
	}
}

// parseForm splits an encoded body without reordering it.
//
// url.ParseQuery returns a map, which loses the order this API cares
// about in places, so the pairs are kept as sent.
func parseForm(body string) []FormPair {
	if body == "" {
		return nil
	}
	parts := strings.Split(body, "&")
	out := make([]FormPair, 0, len(parts))
	for _, part := range parts {
		k, v, _ := strings.Cut(part, "=")
		key := unescape(k)
		val := unescape(v)
		if shouldRedact(key) {
			val = "REDACTED"
		}
		out = append(out, FormPair{Key: key, Value: val})
	}
	return out
}

func unescape(s string) string {
	s = strings.ReplaceAll(s, "+", " ")
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			var n int
			if _, err := fmt.Sscanf(s[i+1:i+3], "%02x", &n); err == nil {
				b.WriteByte(byte(n))
				i += 2
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// redactBody blanks secret-bearing values in a response body.
//
// The request side was redacted from the start and the response side
// was not, which had it exactly backwards: the most sensitive value
// this API ever emits is in a response. server.Create returns the new
// machine's root password, and the package doc notes it "is returned
// once and never again" — so a recording of a provision was the only
// copy of a live credential, sitting in a JSON file, produced by a
// workflow this repository actively recommends.
//
// It walks the decoded body rather than matching text, so a value is
// redacted because of the key it sits under rather than because of
// what it looks like. A body that is not JSON is recorded verbatim:
// this is a diagnostic, and dropping an unparseable response would
// lose exactly the case worth having.
//
// This is a reduction in exposure, not a guarantee. Recordings still
// contain live customer data — server names, addresses, database
// names, usernames, home directories, key material — and must be run
// through [Scrub] before they are committed or shared. That statement
// is the honest one, and it is in the package doc for the same reason
// this function exists: a claim of safety is worth nothing without
// something enforcing it.
func redactBody(raw []byte) string {
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return string(raw)
	}
	out, err := json.Marshal(redactValue(doc))
	if err != nil {
		return string(raw)
	}
	return string(out)
}

// isEmptyValue reports a value that cannot be a secret.
//
// An empty string is not a credential, and redacting one destroys
// evidence: this API never returns a database user's password, and the
// fixture proving that is a run of empty strings. Replacing them with
// REDACTED made the fixture say the opposite, and the test asserting
// the real behaviour failed — correctly.
func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	s, ok := v.(string)
	return ok && s == ""
}

// redactValue walks a decoded body, blanking values under secret keys.
func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if shouldRedact(k) && !isEmptyValue(val) {
				out[k] = "REDACTED"
				continue
			}
			out[k] = redactValue(val)
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, el := range t {
			out = append(out, redactValue(el))
		}
		return out
	default:
		return v
	}
}
