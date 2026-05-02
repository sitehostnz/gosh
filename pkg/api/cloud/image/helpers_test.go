package image

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestCloneURL_Default(t *testing.T) {
	c, _ := api.New("k", "958466", api.SetBaseURL("http://example.invalid"))
	got := New(c).CloneURL("cerb-custom-image")
	want := "git@gitlab-clients.sitehost.co.nz:g_958466/cerb-custom-image.git"
	if got != want {
		t.Errorf("CloneURL = %q, want %q", got, want)
	}
}

func TestCloneURL_OverriddenHost(t *testing.T) {
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"),
		api.SetCustomImageGitHost("gitlab-staging.sitehost.co.nz"))
	got := New(c).CloneURL("test-image")
	if !strings.HasPrefix(got, "git@gitlab-staging.sitehost.co.nz:") {
		t.Errorf("CloneURL = %q, want override host prefix", got)
	}
}

func TestForkFromImage_ResolvesParentID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/cloud/image/list_all.json":
			_, _ = io.WriteString(w, `{
				"status":true,"msg":"Successful",
				"return":{
					"total_items":2,"current_items":2,"current_page":1,"total_pages":1,
					"data":[
						{"id":"61","label":"PHP 8.0","code":"sitehost-php80","is_public":"1"},
						{"id":"99","label":"Mine","code":"my-image","is_public":"0"}
					]
				}
			}`)
		case "/cloud/image/create.json":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "params%5Bfork_id%5D=61") {
				t.Errorf("expected fork_id=61 in body, got: %s", string(body))
			}
			_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"job":{"id":1,"type":"scheduler"}}}`)
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	_, err := New(c).ForkFromImage(context.Background(), "sitehost-php80", "Forked", "forked-img", []int{23})
	if err != nil {
		t.Fatalf("ForkFromImage: %v", err)
	}
}

func TestForkFromImage_ParentNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"OK",
			"return":{"total_items":0,"current_items":0,"current_page":1,"total_pages":1,"data":[]}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	_, err := New(c).ForkFromImage(context.Background(), "nope", "L", "c", []int{1})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestWaitForBuild_TerminatesOnSuccess(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		status := "running"
		if calls >= 2 {
			status = "success"
		}
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{
			"total_items":1,"current_items":1,"current_page":1,"total_pages":1,
			"data":[{"id":"1","client_id":"1","image_id":"77","version":"1.0-1","build_id":"1","build_status":"`+status+`","code":"x","container_count":0}]
		}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).WaitForBuild(context.Background(), 77, 5*time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForBuild: %v", err)
	}
	if got.BuildStatus != "success" {
		t.Errorf("BuildStatus = %q", got.BuildStatus)
	}
}

func TestLintManifest_Valid(t *testing.T) {
	data := []byte(`version: 1
image:
  label: my-custom-image
  type: www
  provider: 'My Account'
`)
	if err := LintManifest(data); err != nil {
		t.Errorf("expected valid, got %v", err)
	}
}

func TestLintManifest_MissingFields(t *testing.T) {
	data := []byte(`version: 2
image:
  label: ""
`)
	err := LintManifest(data)
	if err == nil {
		t.Fatal("expected lint error")
	}
	msg := err.Error()
	for _, want := range []string{"version must be 1", "image.label is required", "image.type is required", "image.provider is required"} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in error: %s", want, msg)
		}
	}
}

func TestLintManifest_BadYAML(t *testing.T) {
	if err := LintManifest([]byte("\tnot: yaml: at all: [")); err == nil {
		t.Error("expected YAML parse error")
	}
}
