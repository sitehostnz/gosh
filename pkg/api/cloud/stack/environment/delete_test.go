package environment

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/environment/delete.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(body))
		if v.Get("server") != "ch-test" {
			t.Errorf("server = %q (note: not 'server_name')", v.Get("server"))
		}
		if v.Get("project") != "myproj" {
			t.Errorf("project = %q", v.Get("project"))
		}
		if v.Get("service") != "web" {
			t.Errorf("service = %q", v.Get("service"))
		}
		if v.Get("variables[0][name]") != "FOO" || v.Get("variables[1][name]") != "BAR" {
			t.Errorf("variables = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":123,"type":"scheduler"}}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Delete(context.Background(), DeleteRequest{
		ServerName: "ch-test", Project: "myproj", Service: "web",
		Names: []string{"FOO", "BAR"},
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.Return.ID != 123 {
		t.Errorf("Return.Job.ID = %d, want 123", got.Return.ID)
	}
	if got.Return.Type != "scheduler" {
		t.Errorf("Return.Job.Type = %q, want \"scheduler\"", got.Return.Type)
	}
}
