package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const testServerName = "ch-test"

func readForm(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	v, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("parse body: %v", err)
	}
	return v
}

func TestGetUpdateWindow_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/server/get_update_window.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != testServerName {
			t.Errorf("server_name = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true, "msg":"Successful",
			"return": {"enabled": true, "day_of_week": 1, "hour_of_day": 3, "minute_of_hour": 0}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).GetUpdateWindow(context.Background(), GetUpdateWindowRequest{ServerName: testServerName})
	if err != nil {
		t.Fatalf("GetUpdateWindow: %v", err)
	}
	if !got.Return.Enabled || got.Return.DayOfWeek != 1 || got.Return.HourOfDay != 3 {
		t.Errorf("Return = %+v", got.Return)
	}
}

const jobOK = `{
	"status":true, "msg":"Successful",
	"return": {"job": {"id": 12345, "type": "scheduler"}}
}`

func TestSetUpdateWindow_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/server/set_update_window.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("server_name") != testServerName {
			t.Errorf("server_name = %q", v.Get("server_name"))
		}
		if v.Get("enabled") != "1" || v.Get("day_of_week") != "2" ||
			v.Get("hour_of_day") != "4" || v.Get("minute_of_hour") != "30" {
			t.Errorf("form = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).SetUpdateWindow(context.Background(), SetUpdateWindowRequest{
		ServerName: testServerName, Enabled: 1, DayOfWeek: 2, HourOfDay: 4, MinuteOfHour: 30,
	})
	if err != nil {
		t.Fatalf("SetUpdateWindow: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestUpdateMinimumTLSVersion_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/server/update_minimum_tls_version.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := readForm(t, r)
		if v.Get("server_name") != testServerName {
			t.Errorf("server_name = %q", v.Get("server_name"))
		}
		if v.Get("minimum_tls_version") != "TLSv1.2" {
			t.Errorf("minimum_tls_version = %q (note 'TLSv1.x' format required)", v.Get("minimum_tls_version"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, jobOK)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpdateMinimumTLSVersion(context.Background(), UpdateMinimumTLSVersionRequest{
		ServerName: testServerName, MinimumTLSVersion: "TLSv1.2",
	})
	if err != nil {
		t.Fatalf("UpdateMinimumTLSVersion: %v", err)
	}
	if got.Return.ID != 12345 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}
