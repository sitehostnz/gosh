package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// bodyOf reads a request body into url.Values.
func bodyOf(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	v, err := url.ParseQuery(string(raw))
	if err != nil {
		t.Fatalf("parse body %q: %v", raw, err)
	}
	return v
}

// TestCreate_DefaultsToAutoIPv4 pins the behaviour callers relied on
// before Params was honoured: no IPv4 given means allocate one.
func TestCreate_DefaultsToAutoIPv4(t *testing.T) {
	t.Parallel()
	var got url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = bodyOf(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"name":"web","ips":["203.0.113.10"]}}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	resp, err := New(c).Create(context.Background(), CreateRequest{
		Label: "web-a", Location: "AKLNCT", ProductCode: "LHPVS1", Image: "ubuntu-2404-20260727",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Auto-allocation goes on the wire as a scalar, params[ipv4]. The
	// bracket form params[ipv4][]=auto is rejected with "The ip
	// address is invalid, please specify a valid ip address", which
	// names an address the caller never supplied and so does not point
	// at the real problem. Explicit addresses do use the bracket form,
	// since more than one may be passed.
	if v := got["params[ipv4]"]; len(v) != 1 || v[0] != "auto" {
		t.Errorf("params[ipv4] = %v, want [auto]", v)
	}
	if v := got["params[ipv4][]"]; len(v) != 0 {
		t.Errorf("params[ipv4][] = %v, want it unset for auto-allocation", v)
	}
	if resp.Return.Name != "web" {
		t.Errorf("Return.Name = %q, want web", resp.Return.Name)
	}
}

// TestCreate_HonoursParams covers the fields that were previously
// dropped silently.
func TestCreate_HonoursParams(t *testing.T) {
	t.Parallel()
	var got url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = bodyOf(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"name":"web"}}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	if _, err := New(c).Create(context.Background(), CreateRequest{
		Label: "web-a", Location: "AKLNCT", ProductCode: "LHPVS1", Image: "img",
		Params: ParamsOptions{
			Name:      "chosen",
			IPv4:      []string{"203.0.113.10", "203.0.113.11"},
			IPv6:      []string{"auto"},
			SSHKeys:   []string{"ssh-ed25519 AAAA a@b"},
			ContactID: "42",
			Backup:    "1",
			SendEmail: "0",
		},
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	for key, want := range map[string][]string{
		"params[ipv4][]":     {"203.0.113.10", "203.0.113.11"},
		"params[ipv6][]":     {"auto"},
		"params[ssh_keys][]": {"ssh-ed25519 AAAA a@b"},
		"params[name]":       {"chosen"},
		"params[contact_id]": {"42"},
		"params[backup]":     {"1"},
		"params[send_email]": {"0"},
	} {
		gotVals := got[key]
		if len(gotVals) != len(want) {
			t.Errorf("%s = %v, want %v", key, gotVals, want)
			continue
		}
		for i := range want {
			if gotVals[i] != want[i] {
				t.Errorf("%s[%d] = %q, want %q", key, i, gotVals[i], want[i])
			}
		}
	}
}

// TestCreate_OmitsUnsetParams checks optional fields are absent rather
// than sent empty.
func TestCreate_OmitsUnsetParams(t *testing.T) {
	t.Parallel()
	var got url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = bodyOf(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"name":"web"}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).Create(context.Background(), CreateRequest{
		Label: "web-a", Location: "AKLNCT", ProductCode: "LHPVS1", Image: "img",
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	for _, key := range []string{"params[name]", "params[ipv6][]", "params[ssh_keys][]", "params[contact_id]", "params[backup]", "params[send_email]"} {
		if _, present := got[key]; present {
			t.Errorf("%s was sent (%v) but should be omitted when unset", key, got[key])
		}
	}
}
