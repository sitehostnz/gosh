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

func TestMount_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/volume/mount.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form["containers[mystack][]"]; len(got) != 1 || got[0] != "container-a" {
			t.Errorf("containers[mystack][] = %v, want [container-a]", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).Mount(context.Background(), MountOptions{
		ServerName: "ch-foo", VolumeName: "data-vol",
		Containers: []ContainerMount{{StackName: "mystack", ContainerName: "container-a"}},
	}); err != nil {
		t.Fatalf("Mount: %v", err)
	}
}

func TestMount_RequiredContainer(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).Mount(context.Background(), MountOptions{
		ServerName: "ch-foo", VolumeName: "data-vol",
	}); err == nil || !strings.Contains(err.Error(), "Container is required") {
		t.Errorf("missing Container: err = %v", err)
	}
}
