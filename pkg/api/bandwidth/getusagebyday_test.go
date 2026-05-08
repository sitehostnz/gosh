package bandwidth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const testIP = "203.0.113.10/32"

func TestGetUsageByDay_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bandwidth/get_usage_by_day.json" {
			t.Errorf("path = %q, want /bandwidth/get_usage_by_day.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("ip_addr"); got != testIP {
			t.Errorf("ip_addr = %q, want %s", got, testIP)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": {
				"203.0.113.10/32": {
					"2026-05-01": {
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

	got, err := New(c).GetUsageByDay(context.Background(), UsageOptions{IPAddr: testIP})
	if err != nil {
		t.Fatalf("GetUsageByDay: %v", err)
	}

	stats := got.Return[testIP]["2026-05-01"]["DOMESTIC"]
	if stats.PeakOut != 209.1016 {
		t.Errorf("DOMESTIC.PeakOut = %v, want 209.1016", stats.PeakOut)
	}
}
