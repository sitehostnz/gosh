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

func TestUpdateMounts_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/volume/update_mounts.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form["containers[add][stackA][]"]; len(got) != 1 || got[0] != "addme" {
			t.Errorf("containers[add][stackA][] = %v, want [addme]", got)
		}
		if got := r.Form["containers[remove][stackB][]"]; len(got) != 1 || got[0] != "removeme" {
			t.Errorf("containers[remove][stackB][] = %v, want [removeme]", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":2,"type":"scheduler"}}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).UpdateMounts(context.Background(), UpdateMountsOptions{
		ServerName: "ch-foo", VolumeName: "data-vol",
		Add:    []ContainerMount{{StackName: "stackA", ContainerName: "addme"}},
		Remove: []ContainerMount{{StackName: "stackB", ContainerName: "removeme"}},
	}); err != nil {
		t.Fatalf("UpdateMounts: %v", err)
	}
}

func TestUpdateMounts_AtLeastOneRequired(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).UpdateMounts(context.Background(), UpdateMountsOptions{
		ServerName: "ch-foo", VolumeName: "data-vol",
	}); err == nil || !strings.Contains(err.Error(), "at least one of Add or Remove") {
		t.Errorf("missing Add/Remove: err = %v", err)
	}
}
