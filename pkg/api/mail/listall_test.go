package mail

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListAll_ParsesUnionTypes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mail/list_all.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("server_name") != "mail-srv" || q.Get("domain") != "example.com" {
			t.Errorf("query = %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful.",
			"return":[
				{"type":0,"username":"bruce@example.com","emailaddr":"bruce@example.com","label":"Bruce"},
				{"type":1,"emailaddr":"thomas@example.com","destination":"bruce@example.com"},
				{"type":2,"emailaddr":"batman@example.com","destination":"bruce@example.com"}
			]
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListAll(context.Background(), ListAllOptions{
		ServerOptions: ServerOptions{ServerName: "mail-srv"},
		Domain:        "example.com",
	})
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(got.Return) != 3 {
		t.Fatalf("expected 3 records, got %d", len(got.Return))
	}
	if got.Return[0].Type != 0 || got.Return[0].Username != "bruce@example.com" {
		t.Errorf("[0] = %+v", got.Return[0])
	}
	if got.Return[1].Type != 1 || got.Return[1].Destination != "bruce@example.com" {
		t.Errorf("[1] = %+v", got.Return[1])
	}
	if got.Return[2].Type != 2 || got.Return[2].EmailAddr != "batman@example.com" {
		t.Errorf("[2] = %+v", got.Return[2])
	}
}

func TestListAll_RequiresServerName(t *testing.T) {
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).ListAll(context.Background(), ListAllOptions{Domain: "x.com"}); err == nil {
		t.Fatal("expected error for missing ServerName")
	}
}

func TestListAll_RequiresDomain(t *testing.T) {
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).ListAll(context.Background(), ListAllOptions{
		ServerOptions: ServerOptions{ServerName: "s"},
	}); err == nil {
		t.Fatal("expected error for missing Domain")
	}
}
