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

	got, err := New(c).ListImages(context.Background())
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
