package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListImages_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/list_images.json" {
			t.Errorf("path = %q, want /server/list_images.json", r.URL.Path)
		}
		// The unfiltered call must not invent filter parameters.
		if got := r.URL.Query().Get("filters[type]"); got != "" {
			t.Errorf("filters[type] = %q, want empty", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{"name": "Debian 13 (Trixie)", "code": "debian-trixie.amd64", "arch": "amd64", "distro": "debian-trixie", "type": "distro", "os": "linux"},
				{"name": "Cloud Server (Trusty)", "code": "cloudhosting-v2-trusty", "arch": "amd64", "distro": "ubuntu-trusty", "type": "salt-container", "os": "linux"}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListImages(context.Background(), ListImagesOptions{})
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}

	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
	if got.Return[0].Code != "debian-trixie.amd64" {
		t.Errorf("Return[0].Code = %q, want debian-trixie.amd64", got.Return[0].Code)
	}
	if got.Return[0].Arch != "amd64" {
		t.Errorf("Return[0].Arch = %q, want amd64", got.Return[0].Arch)
	}
}

// TestListImages_HPVMFiltersOnWire covers the pairing that makes the
// high-performance catalogue reachable at all.
func TestListImages_HPVMFiltersOnWire(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("filters[type]"); got != ImageTypeHPVMDistro {
			t.Errorf("filters[type] = %q, want %q", got, ImageTypeHPVMDistro)
		}
		if got := q.Get("filters[location]"); got != "AKLNCT" {
			t.Errorf("filters[location] = %q, want AKLNCT", got)
		}
		if got := q.Get("filters[include_disabled]"); got != "1" {
			t.Errorf("filters[include_disabled] = %q, want 1", got)
		}
		if got := q.Get("filters[page_size]"); got != "50" {
			t.Errorf("filters[page_size] = %q, want 50", got)
		}
		if got := q.Get("filters[page_number]"); got != "2" {
			t.Errorf("filters[page_number] = %q, want 2", got)
		}
		// Pinned to the wire, not merely passed in the options. A test
		// that only set them would pass identically if filters()
		// stopped emitting them.
		if got := q.Get("filters[sort_by]"); got != "name" {
			t.Errorf("filters[sort_by] = %q, want name", got)
		}
		if got := q.Get("filters[sort_dir]"); got != "desc" {
			t.Errorf("filters[sort_dir] = %q, want desc", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{"name": "Ubuntu 24.04 (Noble)", "code": "ubuntu-2404-20260727", "arch": "amd64", "distro": "ubuntu-noble", "type": "hpvm-distro", "os": "linux"}
			]
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).ListImages(context.Background(), ListImagesOptions{
		Type:            ImageTypeHPVMDistro,
		Location:        "AKLNCT",
		IncludeDisabled: true,
		PageSize:        50,
		PageNumber:      2,
		SortBy:          "name",
		SortDir:         "desc",
	})
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if len(got.Return) != 1 {
		t.Fatalf("len(Return) = %d, want 1", len(got.Return))
	}
	if got.Return[0].Code != "ubuntu-2404-20260727" {
		t.Errorf("Return[0].Code = %q, want ubuntu-2404-20260727", got.Return[0].Code)
	}
}

// TestListImages_HPVMRequiresLocation checks the pairing is enforced
// before a request is made. The API rejects the combination anyway,
// but failing locally gives a clearer message and saves a round trip.
func TestListImages_HPVMRequiresLocation(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("no request should be made when Location is missing")
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	if _, err := New(c).ListImages(context.Background(), ListImagesOptions{
		Type: ImageTypeHPVMDistro,
	}); err == nil {
		t.Fatal("ListImages: expected an error when Type is hpvm-distro and Location is empty")
	}
}
