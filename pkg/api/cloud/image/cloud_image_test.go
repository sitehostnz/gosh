package image

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

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

const jobOK = `{"status":true,"msg":"Successful","return":{"job":{"id":7319505,"type":"scheduler"}}}`

func TestCreate_Forked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/image/create.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("label") != "My Custom Image" || v.Get("params[code]") != "my-custom-image" ||
			v.Get("params[fork_id]") != "61" {
			t.Errorf("form = %v", v)
		}
		if v.Get("params[ssh_keys][0]") != "23" || v.Get("params[ssh_keys][1]") != "25" {
			t.Errorf("ssh_keys = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Create(context.Background(), CreateRequest{
		Label: "My Custom Image", Code: "my-custom-image", ForkID: 61,
		SSHKeys: []int{23, 25},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Return.ID != 7319505 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestCreate_FromScratch_OmitsForkID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := readForm(t, r)
		if _, present := v["params[fork_id]"]; present {
			t.Errorf("fork_id should be omitted when zero, got %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).Create(context.Background(), CreateRequest{Label: "scratch"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestCreate_RequiresLabel(t *testing.T) {
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).Create(context.Background(), CreateRequest{}); err == nil {
		t.Fatal("expected error for missing Label")
	}
}

func TestDelete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/image/delete.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("code") != "my-custom-image" {
			t.Errorf("code = %q", v.Get("code"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Delete(context.Background(), DeleteRequest{Code: "my-custom-image"})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.Return.ID != 7319505 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestGetChangelog_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/image/get_changelog.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("code") != "sitehost-php55" {
			t.Errorf("code = %q", r.URL.Query().Get("code"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful",
			"return":{"code":"sitehost-php55","label":"None","changelog":"**VERSION: 1.0.0**"}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).GetChangelog(context.Background(), GetChangelogRequest{Code: "sitehost-php55"})
	if err != nil {
		t.Fatalf("GetChangelog: %v", err)
	}
	if got.Return.Code != "sitehost-php55" || got.Return.Changelog == "" {
		t.Errorf("Return = %+v", got.Return)
	}
}
