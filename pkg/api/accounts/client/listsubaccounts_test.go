package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const sampleListSubAccounts = `{
  "status": true, "msg": "Successful",
  "return": {
    "total_items": 2, "current_items": 2, "current_page": 1, "total_pages": 1,
    "data": [
      {"client_id":"960001","name":"John Smith","account_balance":"100.12","joined":"2022-07-20","closed":"0000-00-00","account_type":"TOPUP"},
      {"client_id":"960002","name":"Gordon Freeman","account_balance":"41.2","joined":"2004-11-16","closed":"0000-00-00","account_type":"CREDIT"}
    ]
  }
}`

func TestListSubAccounts_Defaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts/client/list_sub_accounts.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		// no filters set — ensure we don't send any
		if got := r.URL.Query().Get("filters[name]"); got != "" {
			t.Errorf("filters[name] = %q (should be unset)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, sampleListSubAccounts)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListSubAccounts(context.Background(), ListSubAccountsRequest{})
	if err != nil {
		t.Fatalf("ListSubAccounts: %v", err)
	}
	if len(got.Return.Accounts) != 2 || got.Return.Accounts[0].ClientID != "960001" {
		t.Errorf("Accounts = %+v", got.Return.Accounts)
	}
}

func TestListSubAccounts_Filters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("filters[name]") != "John" {
			t.Errorf("filters[name] = %q", q.Get("filters[name]"))
		}
		if q.Get("filters[include_closed]") != "1" {
			t.Errorf("filters[include_closed] = %q", q.Get("filters[include_closed]"))
		}
		if q.Get("filters[sort_by]") != "joined" || q.Get("filters[sort_dir]") != "DESC" {
			t.Errorf("sort = %v", q)
		}
		if q.Get("filters[page_size]") != "10" || q.Get("filters[page_number]") != "2" {
			t.Errorf("paging = %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, sampleListSubAccounts)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).ListSubAccounts(context.Background(), ListSubAccountsRequest{
		Name: "John", IncludeClosed: true,
		SortBy: "joined", SortDir: "DESC",
		PageSize: 10, PageNumber: 2,
	}); err != nil {
		t.Fatalf("ListSubAccounts: %v", err)
	}
}
