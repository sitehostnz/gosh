package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestUpgradeComponents_BothFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/upgrade.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(body))
		if v.Get("name") != "myserver" {
			t.Errorf("name = %q", v.Get("name"))
		}
		if v.Get("upgrade[cores]") != "4" || v.Get("upgrade[ram]") != "8" {
			t.Errorf("upgrade fields = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful",
			"return":{"job":{"type":"scheduler","id":7355939},"cores":true,"ram":true}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpgradeComponents(context.Background(), UpgradeComponentsRequest{
		Name: "myserver", Cores: 4, RAM: "8",
	})
	if err != nil {
		t.Fatalf("UpgradeComponents: %v", err)
	}
	if !got.Return.Cores || !got.Return.RAM {
		t.Errorf("expected both cores+ram accepted, got %+v", got.Return)
	}
	if got.Return.ID != 7355939 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestUpgradeComponents_DiskByLabel(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(body))
		if v.Get("upgrade[disk][xvda1]") != "80" {
			t.Errorf("upgrade[disk][xvda1] = %q (want 80)", v.Get("upgrade[disk][xvda1]"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"job":{"id":1,"type":"scheduler"},"disk":true}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpgradeComponents(context.Background(), UpgradeComponentsRequest{
		Name: "s1", Disk: map[string]int{"xvda1": 80},
	}); err != nil {
		t.Fatalf("UpgradeComponents: %v", err)
	}
}

func TestUpgradeComponents_OmitsZeroFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(body))
		if _, present := v["upgrade[cores]"]; present {
			t.Errorf("upgrade[cores] should be omitted when 0, got %v", v)
		}
		if v.Get("upgrade[ram]") != "16" {
			t.Errorf("upgrade[ram] = %q", v.Get("upgrade[ram]"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"job":{"id":1,"type":"scheduler"},"ram":true}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpgradeComponents(context.Background(), UpgradeComponentsRequest{
		Name: "s1", RAM: "16",
	}); err != nil {
		t.Fatalf("UpgradeComponents: %v", err)
	}
}
