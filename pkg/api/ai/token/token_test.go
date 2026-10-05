package token_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/ai/token"
)

// sent is what one request put on the wire.
type sent struct {
	method, path string
	query        url.Values
	form         url.Values
}

// serve answers every request with testdata/<fixture> and records what
// was sent. The fixtures are scrubbed recordings of production
// responses from examples/ai, not hand-written shapes.
func serve(t *testing.T, fixture string) (*token.Client, *sent, []byte) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", fixture)) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatal(err)
	}
	got := &sent{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		got.method, got.path, got.query = r.Method, r.URL.Path, r.URL.Query()
		got.form, _ = url.ParseQuery(string(raw))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return token.New(c), got, body
}

func TestAdd_EncodesCapabilitiesAndAllowlistAsArrays(t *testing.T) {
	t.Parallel()
	tc, got, body := serve(t, "add.json")

	resp, err := tc.Add(context.Background(), token.AddRequest{
		Label:        "l",
		Capabilities: []string{token.CapabilityInference, token.CapabilityEmbedding},
		AllowedIPs:   []string{"192.0.2.1", "198.51.100.0/24"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	apitest.AssertDecodesFully(t, body, token.AddResponse{})

	if got.method != http.MethodPost || got.path != "/ai/token/add.json" {
		t.Errorf("sent %s %s", got.method, got.path)
	}
	if got.form.Get("label") != "l" {
		t.Errorf("label = %q", got.form.Get("label"))
	}
	if c := got.form["capabilities[]"]; !slices.Equal(c, []string{"inference", "embedding"}) {
		t.Errorf("capabilities[] = %v", c)
	}
	if ips := got.form["params[allowed_ips][]"]; !slices.Equal(ips, []string{"192.0.2.1", "198.51.100.0/24"}) {
		t.Errorf("params[allowed_ips][] = %v", ips)
	}
	if resp.Return.KeyID == "" || resp.Return.Token == "" {
		t.Errorf("key_id %q / token %q: the secret is the point of this response", resp.Return.KeyID, resp.Return.Token)
	}
}

func TestAdd_NoAllowlistSendsNoAllowlist(t *testing.T) {
	t.Parallel()
	tc, got, _ := serve(t, "add.json")

	if _, err := tc.Add(context.Background(), token.AddRequest{Label: "l", Capabilities: []string{token.CapabilityInference}}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, ok := got.form["params[allowed_ips][]"]; ok {
		t.Errorf("an unrestricted token sent params[allowed_ips][]=%v", got.form["params[allowed_ips][]"])
	}
}

func TestGet_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	tc, got, body := serve(t, "get.json")

	resp, err := tc.Get(context.Background(), token.GetRequest{KeyID: "a b"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	apitest.AssertDecodesFully(t, body, token.GetResponse{})

	if got.query.Get("key_id") != "a b" {
		t.Errorf("key_id = %q; it must survive query encoding", got.query.Get("key_id"))
	}
	// The fixture is a token with two capabilities and one allowed IP,
	// so each of these can fail on a decode that loses the config.
	c := resp.Return.Config
	if len(c.Capabilities) != 2 || len(c.AllowedIPs) != 1 {
		t.Errorf("config = %+v", c)
	}
	if resp.Return.KeyID == "" || resp.Return.DateAdded == "" {
		t.Errorf("summary fields did not decode: %+v", resp.Return.Summary)
	}
}

func TestGet_RevokedIsAnError(t *testing.T) {
	t.Parallel()
	tc, _, _ := serve(t, "get_revoked.json")

	if _, err := tc.Get(context.Background(), token.GetRequest{KeyID: "x"}); err == nil {
		t.Fatal("Get of a revoked token returned no error")
	}
}

func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	tc, got, body := serve(t, "list_all.json")

	resp, err := tc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	apitest.AssertDecodesFully(t, body, token.ListResponse{})

	if got.path != "/ai/token/list_all.json" {
		t.Errorf("path = %q", got.path)
	}
	if len(resp.Return.Tokens) == 0 {
		t.Fatal("no tokens decoded; the fixture has rows")
	}
	for i, s := range resp.Return.Tokens {
		if s.KeyID == "" || s.Label == "" {
			t.Errorf("Tokens[%d] = %+v", i, s)
		}
	}
}

// The allowlist has three states on an update, and the wire encodings
// differ in a way the API does not report: the scalar
// params[allowed_ips]= is accepted, answers {"updated":true}, and
// changes nothing. These cases pin the encodings observed to work.
func TestUpdate_AllowlistStates(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		ips   []string
		want  []string
		isSet bool
	}{
		{name: "nil leaves it alone", ips: nil, isSet: false},
		{name: "empty clears it", ips: []string{}, want: []string{""}, isSet: true},
		{name: "values replace it", ips: []string{"192.0.2.1"}, want: []string{"192.0.2.1"}, isSet: true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tc, got, _ := serve(t, "update.json")

			if _, err := tc.Update(context.Background(), token.UpdateRequest{KeyID: "k", AllowedIPs: tt.ips}); err != nil {
				t.Fatalf("Update: %v", err)
			}
			v, ok := got.form["params[allowed_ips][]"]
			if ok != tt.isSet || (ok && !slices.Equal(v, tt.want)) {
				t.Errorf("params[allowed_ips][] = %v (present %v), want %v (present %v)", v, ok, tt.want, tt.isSet)
			}
			if _, ok := got.form["params[allowed_ips]"]; ok {
				t.Error("sent the scalar params[allowed_ips], which the API accepts and ignores")
			}
		})
	}
}

func TestUpdate_SendsOnlyWhatIsSet(t *testing.T) {
	t.Parallel()
	tc, got, body := serve(t, "update.json")

	resp, err := tc.Update(context.Background(), token.UpdateRequest{KeyID: "k", Label: "new"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	apitest.AssertDecodesFully(t, body, token.UpdateResponse{})

	if got.path != "/ai/token/update.json" || got.form.Get("key_id") != "k" || got.form.Get("params[label]") != "new" {
		t.Errorf("sent %s %v", got.path, got.form)
	}
	if _, ok := got.form["params[capabilities][]"]; ok {
		t.Error("a label-only update sent capabilities")
	}
	if !resp.Return.Updated {
		t.Error("updated did not decode")
	}
}

func TestDelete_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	tc, got, body := serve(t, "delete.json")

	resp, err := tc.Delete(context.Background(), token.DeleteRequest{KeyID: "k"})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	apitest.AssertDecodesFully(t, body, token.DeleteResponse{})

	if got.method != http.MethodPost || got.path != "/ai/token/delete.json" || got.form.Get("key_id") != "k" {
		t.Errorf("sent %s %s %v", got.method, got.path, got.form)
	}
	if !resp.Return.Deleted {
		t.Error("deleted did not decode")
	}
}

func TestKeyIDIsRequired(t *testing.T) {
	t.Parallel()
	tc, got, _ := serve(t, "update.json")
	ctx := context.Background()

	if _, err := tc.Get(ctx, token.GetRequest{}); err == nil {
		t.Error("Get without a KeyID returned no error")
	}
	if _, err := tc.Update(ctx, token.UpdateRequest{Label: "x"}); err == nil {
		t.Error("Update without a KeyID returned no error")
	}
	if _, err := tc.Delete(ctx, token.DeleteRequest{}); err == nil {
		t.Error("Delete without a KeyID returned no error")
	}
	if got.path != "" {
		t.Errorf("a request went out: %s", got.path)
	}
}
