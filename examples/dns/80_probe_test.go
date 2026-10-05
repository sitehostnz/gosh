package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/dns"
	"github.com/sitehostnz/gosh/pkg/api/dns/template"
)

// TestProbeZoneSkipReason pins the gate that decides whether the probes
// premised on probeZone's absence — two of them a DeleteZone and an
// AddRecord — may run.
//
// The held and absent cases were also run live; the throttle and the
// near-miss search result cannot be produced on demand against
// production, which is why this test exists.
func TestProbeZoneSkipReason(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		status   int
		body     string
		wantSkip bool
		wantErr  bool
	}{
		{
			name:   "absent: an empty search runs the probes",
			status: http.StatusOK,
			body:   `{"status":true,"msg":"Successful","return":[]}`,
		},
		{
			name:     "held: an exact match skips them",
			status:   http.StatusOK,
			body:     `{"status":true,"msg":"Successful","return":[{"name":"` + probeZone + `","client_id":"1","template_id":"0"}]}`,
			wantSkip: true,
		},
		{
			// GetZone is a search. A hit on some other zone is not the
			// probe zone, so it must not be read as "held" — nor, more to
			// the point, does a length check say anything about the name.
			name:   "near miss: a search hit on another name runs the probes",
			status: http.StatusOK,
			body:   `{"status":true,"msg":"Successful","return":[{"name":"x` + probeZone + `","client_id":"1","template_id":"0"}]}`,
		},
		{
			// The API signals a throttle with HTTP 500 and this message.
			// Unable to establish absence, the gate skips rather than
			// failing the step or running.
			name:     "throttled: skipped, not fatal",
			status:   http.StatusInternalServerError,
			body:     `{"msg":"You have exceeded the number of requests per second for this key. Please try again soon."}`,
			wantSkip: true,
		},
		{
			name:    "any other error fails the step",
			status:  http.StatusOK,
			body:    `{"status":false,"msg":"Error: something else"}`,
			wantErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)

			c, err := api.New("k", "1", api.SetBaseURL(srv.URL), api.SetRateLimitRetries(1))
			if err != nil {
				t.Fatal(err)
			}
			skip, err := probeZoneSkipReason(context.Background(), clients{dns: dns.New(c), template: template.New(c)})

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, want error %v", err, tt.wantErr)
			}
			if (skip != "") != tt.wantSkip {
				t.Errorf("skip reason = %q, want skipped %v", skip, tt.wantSkip)
			}
		})
	}
}
