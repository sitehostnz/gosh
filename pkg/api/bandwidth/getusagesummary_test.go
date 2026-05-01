package bandwidth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestGetUsageSummary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bandwidth/get_usage_summary.json" {
			t.Errorf("path = %q, want /bandwidth/get_usage_summary.json", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"203.0.113.10/32": {
					"2026-05": {
						"DOMESTIC":      {"peak_in": 73.4541, "offpeak_in": 0, "peak_out": 209.1016, "offpeak_out": 0},
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

	got, err := New(c).GetUsageSummary(context.Background())
	if err != nil {
		t.Fatalf("GetUsageSummary: %v", err)
	}

	stats := got.Return["203.0.113.10/32"]["2026-05"]["INTERNATIONAL"]
	if stats.PeakIn != 6370.6787 {
		t.Errorf("INTERNATIONAL.PeakIn = %v, want 6370.6787", stats.PeakIn)
	}
	if stats.PeakOut != 7024.5947 {
		t.Errorf("INTERNATIONAL.PeakOut = %v, want 7024.5947", stats.PeakOut)
	}
}
