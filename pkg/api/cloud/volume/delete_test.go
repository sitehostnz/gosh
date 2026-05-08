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

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/volume/delete.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("server"); got != testServerName {
			t.Errorf("server = %q (note: not server_name)", got)
		}
		if got := r.Form.Get("volume"); got != testVolumeName {
			t.Errorf("volume = %q (note: not volume_name)", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":42,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).Delete(context.Background(), DeleteOptions{Server: testServerName, Volume: testVolumeName}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDelete_RequiredFields(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).Delete(context.Background(), DeleteOptions{}); err == nil ||
		!strings.Contains(err.Error(), "Server is required") {
		t.Errorf("missing Server: err = %v", err)
	}
}
