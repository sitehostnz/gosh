package volume

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestList_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/volume/list_all.json" {
			t.Errorf("path = %q, want /cloud/volume/list_all.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"current_items": 1,
				"current_page": 1,
				"total_pages": 1,
				"total_items": 1,
				"data": [
					{
						"id": "100", "client_id": "1234", "server_id": "62994",
						"pending": "", "volume_name": "data-vol",
						"is_missing": "0",
						"date_added": "2026-05-01 00:00:00", "date_updated": "2026-05-02 00:00:00",
						"server_name": "ch-foo", "server_label": "foo",
						"server_owner": true,
						"containers": ["cc1234"]
					}
				]
			}
		}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	got, err := New(c).List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Return.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(got.Return.Data))
	}
	v := got.Return.Data[0]
	if v.VolumeName != "data-vol" {
		t.Errorf("VolumeName = %q", v.VolumeName)
	}
	if !v.ServerOwner {
		t.Errorf("ServerOwner = false, want true")
	}
	if len(v.Containers) != 1 || v.Containers[0] != "cc1234" {
		t.Errorf("Containers = %v, want [cc1234]", v.Containers)
	}
}

func TestList_Filters(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("filters[server_name]"); got != "ch-foo" {
			t.Errorf("filters[server_name] = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"data":[]}}`)
	}))
	defer server.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(server.URL))
	if _, err := New(c).List(context.Background(), &ListOptions{ServerName: "ch-foo"}); err != nil {
		t.Fatalf("List: %v", err)
	}
}
