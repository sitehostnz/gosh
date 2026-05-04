package version

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

func TestListAll_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/image/version/list_all.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("image_id") != "77" {
			t.Errorf("image_id = %q", r.URL.Query().Get("image_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful",
			"return":{
				"total_items":1,"current_items":1,"current_page":1,"total_pages":1,
				"data":[{
					"id":"6421","client_id":"1","image_id":"415","version":"1.1-1076",
					"build_id":"1076","build_status":"success","code":"my-custom-image",
					"container_count":0
				}]
			}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListAll(context.Background(), ListAllRequest{ImageID: 77})
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(got.Return.Versions) != 1 || got.Return.Versions[0].BuildStatus != "success" {
		t.Errorf("Versions = %+v", got.Return.Versions)
	}
}

func TestListAll_RequiresImageID(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).ListAll(context.Background(), ListAllRequest{}); err == nil {
		t.Fatal("expected error for missing ImageID")
	}
}

func TestGetBuild_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/image/version/get_build.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("code") != "my-custom-image" || q.Get("build_id") != "11" {
			t.Errorf("query = %v", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful",
			"return":{"build_status":"failed","build_trace":"<build log here>"}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).GetBuild(context.Background(),
		GetBuildRequest{Code: "my-custom-image", BuildID: "11"})
	if err != nil {
		t.Fatalf("GetBuild: %v", err)
	}
	if got.Return.BuildStatus != "failed" || got.Return.BuildTrace == "" {
		t.Errorf("Return = %+v", got.Return)
	}
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/image/version/delete.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("code") != "my-custom-image" || v.Get("version") != "1.1-1076" {
			t.Errorf("form = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful",
			"return":{"job":{"type":"scheduler","id":7319505}}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Delete(context.Background(),
		DeleteRequest{Code: "my-custom-image", Version: "1.1-1076"})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.Return.ID != 7319505 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}
