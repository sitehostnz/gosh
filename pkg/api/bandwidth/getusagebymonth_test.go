package bandwidth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGetUsageByMonth_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bandwidth/get_usage_by_month.json" {
			t.Errorf("path = %q, want /bandwidth/get_usage_by_month.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("ip_addr"); got != "203.0.113.10/32" {
			t.Errorf("ip_addr = %q, want 203.0.113.10/32", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"203.0.113.10/32": {
					"2026-05": {
						"INTERNATIONAL": {"peak_in": 6370.6787, "offpeak_in": 0, "peak_out": 7024.5947, "offpeak_out": 0}
					}
				}
			}
		}`)
	}))
	defer server.Close()

	c, err := api.New("k", "1", api.SetBaseURL(server.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).GetUsageByMonth(context.Background(), UsageOptions{IPAddr: "203.0.113.10/32"})
	if err != nil {
		t.Fatalf("GetUsageByMonth: %v", err)
	}

	if _, ok := got.Return["203.0.113.10/32"]["2026-05"]; !ok {
		t.Fatal("expected 2026-05 period key")
	}
}
