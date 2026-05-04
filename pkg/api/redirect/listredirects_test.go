package redirect

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListRedirects_ParsesNestedShape(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/redirect/list_redirects.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true, "msg": "Successful",
			"return": {
				"example.kiwi.nz": {
					"a.example.kiwi.nz": {"destination": "https://movedpermanently.com", "type": 301},
					"b.example.kiwi.nz/blog/category": {"destination": "https://movedtemporarily.com", "type": 302}
				}
			}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListRedirects(context.Background(), ListRedirectsRequest{})
	if err != nil {
		t.Fatalf("ListRedirects: %v", err)
	}
	rules := got.Return["example.kiwi.nz"]
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules under example.kiwi.nz, got %d", len(rules))
	}
	if rules["a.example.kiwi.nz"].Type != 301 {
		t.Errorf("a.example.kiwi.nz type = %d", rules["a.example.kiwi.nz"].Type)
	}
	if rules["b.example.kiwi.nz/blog/category"].Destination != "https://movedtemporarily.com" {
		t.Errorf("destination = %q", rules["b.example.kiwi.nz/blog/category"].Destination)
	}
}

func TestListRedirects_EmptyArrayShape(t *testing.T) {
	t.Parallel()
	// The live API returns the JSON array [] (not the empty object
	// {}) when an account has no redirects. Custom UnmarshalJSON
	// must tolerate this.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"return":[],"msg":"Successful","status":true}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListRedirects(context.Background(), ListRedirectsRequest{})
	if err != nil {
		t.Fatalf("ListRedirects: %v", err)
	}
	if got.Return == nil {
		t.Errorf("Return should be non-nil even for empty result")
	}
	if len(got.Return) != 0 {
		t.Errorf("Return = %+v, want empty map", got.Return)
	}
}

func TestListRedirects_PagingFilters(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("filters[page_size]") != "50" || q.Get("filters[page_number]") != "3" {
			t.Errorf("paging = %v", q)
		}
		if q.Get("filters[sort_by]") != "domain" || q.Get("filters[sort_dir]") != "ASC" {
			t.Errorf("sort = %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).ListRedirects(context.Background(), ListRedirectsRequest{
		SortBy: "domain", SortDir: "ASC", PageSize: 50, PageNumber: 3,
	}); err != nil {
		t.Fatalf("ListRedirects: %v", err)
	}
}
