// Package apitest serves recorded API responses to the SDK's clients.
//
// The fixtures under each package's testdata directory are real
// responses from api.sitehost.nz, scrubbed by [recorder.Scrub] so that
// every value is a placeholder and every shape is the one the API
// actually sent. Tests built on them therefore assert the wire, not our
// assumptions about it — which is the distinction that matters, because
// two bugs in this SDK survived a green suite built on hand-written
// mocks that encoded the assumption instead.
//
// Nothing here talks to the network.
package apitest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sitehostnz/gosh/internal/shapecheck"
	"github.com/sitehostnz/gosh/pkg/api"
)

// Exchange is a served fixture and the request that fetched it.
type Exchange struct {
	// Client is wired to the test server; use it to build the SDK
	// client under test.
	Client *api.Client

	// Request is the last request the server received, so a test can
	// check what went on the wire as well as what came back.
	Request *http.Request

	// Body is the raw fixture, for [AssertDecodesFully].
	Body []byte
}

// Serve reads testdata/<fixture> and serves it for every request,
// returning an API client pointed at it.
//
// The fixture is served whatever the path, because these tests are
// about decoding a known response rather than about routing. Check the
// path with [Exchange.Request] where it matters.
func Serve(t *testing.T, fixture string) *Exchange {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("testdata", fixture)) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("apitest: %v", err)
	}

	ex := &Exchange{Body: body}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ex.Request = r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json")
		// This API answers 200 even when it rejects, so the fixture's
		// own envelope is what decides success. Serving a rejection
		// with a 4xx here would test a code path that does not exist.
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("apitest: api.New: %v", err)
	}
	ex.Client = c
	return ex
}

// AssertDecodesFully fails if the fixture contains any field that v has
// no place to put.
//
// encoding/json drops what it does not recognise without a word, so a
// decode can succeed while losing most of the response. This is the
// assertion that catches it, and it is worth making on every fixture:
// it is the one check that gets stronger as the API grows, because a
// field added upstream shows up here as a failure rather than as a
// value that quietly never arrives.
func AssertDecodesFully(t *testing.T, body []byte, v any) {
	t.Helper()
	missing, err := shapecheck.Undecoded(body, v)
	if err != nil {
		t.Fatalf("apitest: %v", err)
	}
	for _, path := range missing {
		t.Errorf("the API sends %s and no field decodes it; it is being dropped silently", path)
	}
}

// ReadFixture returns a fixture's bytes.
func ReadFixture(t *testing.T, fixture string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", fixture)) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("apitest: %v", err)
	}
	return b
}

// Discard drains and closes a body.
func Discard(r io.ReadCloser) {
	_, _ = io.Copy(io.Discard, r)
	_ = r.Close()
}
