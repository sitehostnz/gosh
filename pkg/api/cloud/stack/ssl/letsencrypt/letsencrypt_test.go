package letsencrypt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const jobOK = `{
	"status": true, "msg": "Successful.",
	"return": {"job": {"id": 12345, "type": "scheduler"}}
}`

func readForm(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	v, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("parse body: %v", err)
	}
	return v
}

func TestList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/ssl/lets_encrypt/list_all.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("method = %q", r.Method)
		}
		if got := r.URL.Query().Get("server"); got != "ch-test" {
			t.Errorf("server = %q (note: not server_name)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful.",
			"return": {
				"cc1234": {
					"issuer": "E8",
					"not_before": "2026-01-01 00:00:00",
					"not_after":  "2026-04-01 00:00:00",
					"serial":     "1234567890",
					"expired":    "0",
					"is_missing": "0"
				}
			}
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).List(context.Background(), ListRequest{ServerName: "ch-test"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d", len(got.Return))
	}
	cert, ok := got.Return["cc1234"]
	if !ok {
		t.Fatalf("Return missing cc1234 key")
	}
	if cert.Issuer != "E8" {
		t.Errorf("Issuer = %q", cert.Issuer)
	}
	if cert.Expired != "0" {
		t.Errorf("Expired = %q (string-typed bool)", cert.Expired)
	}
}

func assertWriteSends(t *testing.T, path, server, name string, fn func(*Client) error) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Errorf("path = %q, want %q", r.URL.Path, path)
		}
		if r.Method != "POST" {
			t.Errorf("method = %q", r.Method)
		}
		v := readForm(t, r)
		if v.Get("server") != server {
			t.Errorf("server = %q", v.Get("server"))
		}
		if v.Get("name") != name {
			t.Errorf("name = %q", v.Get("name"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err := fn(New(c)); err != nil {
		t.Fatalf("call: %v", err)
	}
}

func TestCreate_Success(t *testing.T) {
	assertWriteSends(t, "/cloud/stack/ssl/lets_encrypt/create.json", "ch-test", "cc1234", func(c *Client) error {
		got, err := c.Create(context.Background(), CreateRequest{ServerName: "ch-test", Name: "cc1234"})
		if err != nil {
			return err
		}
		if got.Return.ID != 12345 {
			t.Errorf("Job.ID = %d", got.Return.ID)
		}
		return nil
	})
}

func TestDelete_Success(t *testing.T) {
	assertWriteSends(t, "/cloud/stack/ssl/lets_encrypt/delete.json", "ch-test", "cc1234", func(c *Client) error {
		_, err := c.Delete(context.Background(), DeleteRequest{ServerName: "ch-test", Name: "cc1234"})
		return err
	})
}

func TestRenew_Success(t *testing.T) {
	assertWriteSends(t, "/cloud/stack/ssl/lets_encrypt/renew.json", "ch-test", "cc1234", func(c *Client) error {
		_, err := c.Renew(context.Background(), RenewRequest{ServerName: "ch-test", Name: "cc1234"})
		return err
	})
}

func TestRevoke_Success(t *testing.T) {
	assertWriteSends(t, "/cloud/stack/ssl/lets_encrypt/revoke.json", "ch-test", "cc1234", func(c *Client) error {
		_, err := c.Revoke(context.Background(), RevokeRequest{ServerName: "ch-test", Name: "cc1234"})
		return err
	})
}
