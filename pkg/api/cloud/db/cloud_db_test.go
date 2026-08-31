package db_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
)

// TestList_DecodesARecordedResponse checks the paged listing against a
// real response rather than an invented one.
func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all.json")

	got, err := db.New(ex.Client).List(context.Background(), db.ListOptions{ServerName: "s"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, db.ListResponse{})

	if !got.Status {
		t.Fatalf("Status = false, msg = %q", got.Msg)
	}
	if len(got.Return.Databases) == 0 {
		t.Fatal("no databases decoded; the fixture has rows")
	}
	if got.Return.TotalItems == 0 {
		t.Error("TotalItems = 0; the pagination envelope did not decode")
	}

	// Every id on this API is a quoted integer. A test that let it be
	// a bare number would pass against a mock and fail against the API.
	for i, d := range got.Return.Databases {
		if d.ID == "" {
			t.Errorf("Databases[%d].ID is empty", i)
		}
		if d.MySQLHost == "" {
			t.Errorf("Databases[%d].MySQLHost is empty; it is the value db.Get needs", i)
		}
	}

	// The filter has to reach the wire, or an empty result would be
	// indistinguishable from a filter that was silently ignored.
	if got := ex.Request.URL.Query().Get("filters[server_name]"); got != "s" {
		t.Errorf("filters[server_name] = %q, want s", got)
	}
}

// TestList_RejectsAnUnknownServer records the behaviour that makes the
// filter trustworthy: a server name that does not resolve is an error,
// never an empty page.
func TestList_RejectsAnUnknownServer(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_all-unknown-server.json")

	_, err := db.New(ex.Client).List(context.Background(), db.ListOptions{ServerName: "nope"})
	if err == nil {
		t.Fatal("List: expected an error; the API rejects an unresolvable server name")
	}
}

// TestGet_DecodesARecordedResponse covers the single-database read, and
// with it the fix for the duplicated client_id this endpoint used to
// send.
func TestGet_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "get.json")

	got, err := db.New(ex.Client).Get(context.Background(), db.GetRequest{
		ServerName: "s", MySQLHost: "h", Database: "d",
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, db.GetResponse{})

	if got.Database.DBName == "" {
		t.Error("DBName is empty; the single-database body did not decode")
	}

	q := ex.Request.URL.Query()
	if len(q["client_id"]) != 1 {
		t.Errorf("client_id appears %d times, want 1 — it was being added twice", len(q["client_id"]))
	}
	if q.Get("api_key") != "" {
		t.Error("api_key is on the query; the parameter is apikey and this one was silently ignored")
	}
	for _, k := range []string{"server_name", "mysql_host", "database"} {
		if q.Get(k) == "" {
			t.Errorf("%s is missing from the query", k)
		}
	}
}
