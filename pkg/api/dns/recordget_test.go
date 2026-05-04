package dns

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/models"
)

// listRecordsHandler returns a fixed three-record response: an A
// record at id=2, a CNAME at id=3, and an A record at id=4. Used
// by all GetRecord* tests since each is a client-side filter on
// ListRecords output.
func listRecordsHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/list_records.json" {
			t.Errorf("path = %q, want /dns/list_records.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{"id": "2", "name": "www.example.co.nz",  "type": "A",     "content": "192.0.2.10", "ttl": "3600", "prio": "0", "change_date": "1700000001", "state": "0"},
				{"id": "3", "name": "blog.example.co.nz", "type": "CNAME", "content": "www.example.co.nz", "ttl": "3600", "prio": "0", "change_date": "1700000002", "state": "0"},
				{"id": "4", "name": "api.example.co.nz",  "type": "A",     "content": "192.0.2.20", "ttl": "3600", "prio": "0", "change_date": "1700000003", "state": "0"}
			]
		}`)
	}
}

func TestGetRecord_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(listRecordsHandler(t))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetRecord(context.Background(), RecordRequest{
		ID: "3", DomainName: testDomain,
	})
	if err != nil {
		t.Fatalf("GetRecord: %v", err)
	}

	if got.ID != "3" {
		t.Errorf("ID = %q, want 3", got.ID)
	}
	if got.Type != "CNAME" {
		t.Errorf("Type = %q, want CNAME", got.Type)
	}
}

func TestGetRecordWithType_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(listRecordsHandler(t))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetRecordWithType(context.Background(), RecordRequest{
		RRType: "A", DomainName: testDomain,
	})
	if err != nil {
		t.Fatalf("GetRecordWithType: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("len(got) = %d, want 2 A records", len(got))
	}
	for _, r := range got {
		if r.Type != "A" {
			t.Errorf("filtered record has Type = %q, want A", r.Type)
		}
	}
}

func TestGetRecordWithRecord_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(listRecordsHandler(t))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetRecordWithRecord(context.Background(), models.DNSRecord{
		Domain:   testDomain,
		Name:     "www.example.co.nz",
		Type:     "A",
		Content:  "192.0.2.10",
		Priority: "0",
	})
	if err != nil {
		t.Fatalf("GetRecordWithRecord: %v", err)
	}

	if got.ID != "2" {
		t.Errorf("ID = %q, want 2", got.ID)
	}
}
