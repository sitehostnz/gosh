package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListRecords_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/list_records.json" {
			t.Errorf("path = %q, want /dns/list_records.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{"id": "1", "name": "example.co.nz", "type": "NS", "content": "ns1.sitehost.co.nz.", "ttl": "3600", "prio": "0", "change_date": "1700000000", "state": "0"},
				{"id": "2", "name": "www.example.co.nz", "type": "A", "content": "192.0.2.10", "ttl": "3600", "prio": "0", "change_date": "1700000001", "state": "0"}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListRecords(context.Background(), ListRecordsRequest{Domain: testDomain})
	if err != nil {
		t.Fatalf("ListRecords: %v", err)
	}

	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
	if got.Return[1].Type != "A" {
		t.Errorf("Return[1].Type = %q, want A", got.Return[1].Type)
	}
	if got.Return[1].Content != "192.0.2.10" {
		t.Errorf("Return[1].Content = %q, want 192.0.2.10", got.Return[1].Content)
	}
}
