package volume

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const (
	testServerName = "ch-foo"
	testVolumeName = "data-vol"
)

func TestAdd_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/cloud/volume/add.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("server_name"); got != testServerName {
			t.Errorf("server_name = %q", got)
		}
		if got := r.Form.Get("volume_name"); got != testVolumeName {
			t.Errorf("volume_name = %q", got)
		}
		if got := r.Form["container_names[]"]; len(got) != 2 || got[0] != "cc1" || got[1] != "cc2" {
			t.Errorf("container_names[] = %v, want [cc1 cc2]", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":29658430,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).Add(context.Background(), AddOptions{
		ServerName: testServerName, VolumeName: testVolumeName,
		ContainerNames: []string{"cc1", "cc2"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got.Return.ID != 29658430 {
		t.Errorf("Job.ID = %d, want 29658430", got.Return.ID)
	}
}

func TestAdd_RequiredFields(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).Add(context.Background(), AddOptions{VolumeName: "v"}); err == nil ||
		!strings.Contains(err.Error(), "ServerName is required") {
		t.Errorf("missing ServerName: err = %v", err)
	}
	if _, err := New(c).Add(context.Background(), AddOptions{ServerName: "s"}); err == nil ||
		!strings.Contains(err.Error(), "VolumeName is required") {
		t.Errorf("missing VolumeName: err = %v", err)
	}
}
