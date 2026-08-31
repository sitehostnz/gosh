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
		// disk answers per label. This fixture previously said
		// "disk":true, which is not what the API sends — see the
		// UpgradeComponentsResponse doc comment.
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"job":{"id":1,"type":"scheduler"},"disk":{"xvda1":true}}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpgradeComponents(context.Background(), UpgradeComponentsRequest{
		Name: "s1", Disk: map[string]int{"xvda1": 80},
	})
	if err != nil {
		t.Fatalf("UpgradeComponents: %v", err)
	}
	if !got.Return.Disk["xvda1"] {
		t.Errorf("Return.Disk[xvda1] = %v, want true (got %v)", got.Return.Disk["xvda1"], got.Return.Disk)
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

// TestUpgradeComponents_DiskIsKeyedByLabel covers the response shape
// that previously failed to decode: disk answers per label, not as a
// single bool.
func TestUpgradeComponents_DiskIsKeyedByLabel(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"disk":{"scsi0":true}}}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	got, err := New(c).UpgradeComponents(context.Background(), UpgradeComponentsRequest{
		Name: "srv", Disk: map[string]int{"scsi0": 60},
	})
	if err != nil {
		t.Fatalf("UpgradeComponents: %v", err)
	}
	if !got.Return.Disk["scsi0"] {
		t.Errorf("Return.Disk[scsi0] = %v, want true (got map %v)", got.Return.Disk["scsi0"], got.Return.Disk)
	}
	// No job is returned when the resize is applied inline.
	if got.Return.ID != 0 {
		t.Errorf("Return.ID = %d, want 0 for an inline resize", got.Return.ID)
	}
}

// TestUpgradeComponents_ScalarDiskAcceptance covers the shape
// shtypes.MaybeBoolMap was added to tolerate, from the caller's side.
//
// A type-side test proves the value decodes. It does not prove a caller
// reads it correctly, and the one caller in this repository was
// indexing the map directly — so the scalar form, which decodes to an
// empty non-nil map, read as a rejection. That is the shape the type
// exists for, misread by its only consumer.
func TestUpgradeComponents_ScalarDiskAcceptance(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		body     string
		accepted bool
		perKey   bool
	}{
		{
			name:     "the per-disk object form",
			body:     `{"status":true,"msg":"Successful","return":{"disk":{"scsi0":true},"job":{"id":1,"type":"scheduler"}}}`,
			accepted: true,
			perKey:   true,
		},
		{
			name:     "a bare true is an acceptance with nothing to enumerate",
			body:     `{"status":true,"msg":"Successful","return":{"disk":true,"job":{"id":1,"type":"scheduler"}}}`,
			accepted: true,
		},
		{
			name:     "the empty-map form PHP writes as a list",
			body:     `{"status":true,"msg":"Successful","return":{"disk":[],"job":{"id":1,"type":"scheduler"}}}`,
			accepted: true,
		},
		{
			name: "a bare false is not an acceptance",
			body: `{"status":true,"msg":"Successful","return":{"disk":false}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()

			c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
			if err != nil {
				t.Fatalf("api.New: %v", err)
			}
			got, err := New(c).UpgradeComponents(context.Background(), UpgradeComponentsRequest{
				Name: "s", Disk: map[string]int{"scsi0": 100},
			})
			if err != nil {
				t.Fatalf("UpgradeComponents: %v", err)
			}
			if got.Return.Disk.Accepted() != tc.accepted {
				t.Errorf("Disk.Accepted() = %t, want %t", got.Return.Disk.Accepted(), tc.accepted)
			}
			if got.Return.Disk.AcceptedKey("scsi0") != tc.perKey {
				t.Errorf("Disk.AcceptedKey(scsi0) = %t, want %t", got.Return.Disk.AcceptedKey("scsi0"), tc.perKey)
			}
		})
	}
}
