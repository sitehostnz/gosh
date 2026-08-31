// Package cloud_test asserts what the cloud clients put on the wire.
//
// These are request-shape tests, and they exist because the request
// side of this SDK has a silent failure mode. net.Encode emits only the
// keys named in its keys list, so a parameter added under a name absent
// from that list is dropped without an error — the call succeeds, the
// value never arrives, and nothing anywhere says so.
//
// Three bugs of exactly that kind were found by reading the wire rather
// than the code, and each is pinned below.
package cloud_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
)

// sent captures the parameters of the request a call produces, from
// the query string and the form body alike.
func sent(t *testing.T, call func(*api.Client) error) url.Values {
	t.Helper()

	got := url.Values{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		for k, v := range r.Form {
			got[k] = append(got[k], v...)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":true,"msg":"Successful","return":{}}`))
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	if err := call(c); err != nil {
		t.Fatalf("call: %v", err)
	}
	return got
}

// TestDBUserGet_SendsEachCredentialOnce pins the bug that was found by
// reading a recorded URL: client_id appeared twice, and an "api_key"
// parameter this API does not have was being added and then silently
// dropped by net.Encode for not being in the keys list.
func TestDBUserGet_SendsEachCredentialOnce(t *testing.T) {
	t.Parallel()
	got := sent(t, func(c *api.Client) error {
		_, err := dbuser.New(c).Get(context.Background(), dbuser.GetRequest{
			ServerName: "s", MySQLHost: "h", Username: "u",
		})
		return err
	})

	if n := len(got["client_id"]); n != 1 {
		t.Errorf("client_id sent %d times, want 1 — NewRequest already adds it", n)
	}
	if n := len(got["apikey"]); n != 1 {
		t.Errorf("apikey sent %d times, want 1", n)
	}
	if _, present := got["api_key"]; present {
		t.Error("api_key is on the wire; the parameter is apikey and this one was only ever dropped")
	}
	for _, k := range []string{"server_name", "mysql_host", "username"} {
		if got.Get(k) == "" {
			t.Errorf("%s is missing", k)
		}
	}
}

// TestSSHUserUpdate_SendsReadOnlyConfig pins a field that was never
// sent at all.
//
// The value was added as params[read_only_config] while the keys list
// named params[read_only_config][] — with the array suffix — so
// net.Encode skipped it. Setting ReadOnlyConfig was a no-op, and the
// call still succeeded, so nothing indicated it had been ignored.
func TestSSHUserUpdate_SendsReadOnlyConfig(t *testing.T) {
	t.Parallel()
	got := sent(t, func(c *api.Client) error {
		_, err := sshuser.New(c).Update(context.Background(), sshuser.UpdateRequest{
			ServerName:     "s",
			Username:       "u",
			ReadOnlyConfig: true,
		})
		return err
	})

	if v := got.Get("params[read_only_config]"); v == "" {
		t.Errorf("params[read_only_config] was not sent; the field is silently ignored. sent: %v", got)
	}
	if _, present := got["params[read_only_config][]"]; present {
		t.Error("params[read_only_config][] is on the wire; the API takes the scalar form")
	}
}

// TestDBAddAndDelete_SendDatabaseOnce pins a duplicate: the database
// name was added twice, so it went out twice.
func TestDBAddAndDelete_SendDatabaseOnce(t *testing.T) {
	t.Parallel()

	t.Run("add", func(t *testing.T) {
		t.Parallel()
		got := sent(t, func(c *api.Client) error {
			_, err := db.New(c).Add(context.Background(), db.AddRequest{
				ServerName: "s", MySQLHost: "h", Database: "d", Container: "c",
			})
			return err
		})
		if n := len(got["database"]); n != 1 {
			t.Errorf("database sent %d times, want 1", n)
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		got := sent(t, func(c *api.Client) error {
			_, err := db.New(c).Delete(context.Background(), db.DeleteRequest{
				ServerName: "s", MySQLHost: "h", Database: "d",
			})
			return err
		})
		if n := len(got["database"]); n != 1 {
			t.Errorf("database sent %d times, want 1", n)
		}
	})
}
