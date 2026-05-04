package stack

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const (
	testServerName = "srv1"
	testStackName  = "stack1"
)

// jobOK returns a stock JobResponse body for write-write tests.
const jobOK = `{
	"status": true, "msg": "Successful.",
	"return": {"job": {"id": 12345, "type": "scheduler"}}
}`

// readForm parses the POST body of a request as form values.
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

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/update.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("method = %q", r.Method)
		}
		v := readForm(t, r)
		if v.Get("server") != testServerName {
			t.Errorf("server = %q", v.Get("server"))
		}
		if v.Get("name") != testStackName {
			t.Errorf("name = %q", v.Get("name"))
		}
		if v.Get("label") != "new-label" {
			t.Errorf("label = %q", v.Get("label"))
		}
		if v.Get("enable_ssl") != "1" {
			t.Errorf("enable_ssl = %q", v.Get("enable_ssl"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Update(context.Background(), UpdateRequest{
		ServerName: testServerName, Name: testStackName, Label: "new-label", EnableSSL: 1,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/delete.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("server") != testServerName || v.Get("name") != testStackName {
			t.Errorf("server=%q name=%q", v.Get("server"), v.Get("name"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Delete(context.Background(), DeleteRequest{
		ServerName: testServerName, Name: testStackName,
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestCopy_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/copy.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("source_server") != testServerName {
			t.Errorf("source_server = %q", v.Get("source_server"))
		}
		if v.Get("name") != testStackName {
			t.Errorf("name = %q", v.Get("name"))
		}
		if v.Get("destination_server") != "srv2" {
			t.Errorf("destination_server = %q", v.Get("destination_server"))
		}
		if v.Get("label") != "copy-label" {
			t.Errorf("label = %q", v.Get("label"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Copy(context.Background(), CopyRequest{
		SourceServer: testServerName, Name: testStackName,
		DestinationServer: "srv2", Label: "copy-label",
	})
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestOverwrite_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/overwrite.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("source_server") != testServerName {
			t.Errorf("source_server = %q", v.Get("source_server"))
		}
		if v.Get("name") != testStackName {
			t.Errorf("name = %q", v.Get("name"))
		}
		if v.Get("destination_server") != "srv2" {
			t.Errorf("destination_server = %q", v.Get("destination_server"))
		}
		if v.Get("destination_stack") != "stack2" {
			t.Errorf("destination_stack = %q (note: not destination_name)", v.Get("destination_stack"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Overwrite(context.Background(), OverwriteRequest{
		SourceServer: testServerName, Name: testStackName,
		DestinationServer: "srv2", DestinationStack: "stack2",
	})
	if err != nil {
		t.Fatalf("Overwrite: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestBackup_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/backup.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("server") != testServerName || v.Get("name") != testStackName {
			t.Errorf("server=%q name=%q", v.Get("server"), v.Get("name"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Backup(context.Background(), BackupRequest{
		ServerName: testServerName, Name: testStackName,
	})
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestPurgeCache_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/purge_cache.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("server") != testServerName || v.Get("name") != testStackName {
			t.Errorf("server=%q name=%q", v.Get("server"), v.Get("name"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).PurgeCache(context.Background(), PurgeCacheRequest{
		ServerName: testServerName, Name: testStackName,
	})
	if err != nil {
		t.Fatalf("PurgeCache: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestWrites_StatusFalse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status": false, "msg": "Stack not found"}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	_, err := New(c).Delete(context.Background(), DeleteRequest{ServerName: testServerName, Name: "missing"})
	if err == nil {
		t.Fatal("expected error on status:false, got nil")
	}
	if !strings.Contains(err.Error(), "Stack not found") {
		t.Errorf("err = %v", err)
	}
}
