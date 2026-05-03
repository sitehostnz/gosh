package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func TestListIPs_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/list_ips.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("location"); got != "AKLCITY" {
			t.Errorf("location = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		// The actual API returns objects, not bare strings —
		// {ip_addr, prefix, family} per IP. The earlier
		// `Return []string` shape silently dropped this.
		_, _ = io.WriteString(w, `{
			"status":true, "msg":"Successful",
			"return":[
				{"ip_addr":"203.0.113.1","prefix":32,"family":4},
				{"ip_addr":"203.0.113.2","prefix":32,"family":4}
			]}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListIPs(context.Background(), ListIPsOptions{Location: "AKLCITY"})
	if err != nil {
		t.Fatalf("ListIPs: %v", err)
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d", len(got.Return))
	}
	if got.Return[0].IPAddr != "203.0.113.1" || got.Return[0].Family != 4 {
		t.Errorf("first entry = %+v", got.Return[0])
	}
}

func TestListAllocatedIPs_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note the literal underscore in the path — list_allocated_i_ps.
		if r.URL.Path != "/server/list_allocated_i_ps.json" {
			t.Errorf("path = %q (note literal underscore in i_ps)", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true, "msg":"Successful",
			"return": {
				"203.0.113.10": {"ip_addr":"203.0.113.10","netmask":"255.255.255.0","gateway":"203.0.113.1","location":"AKLCITY","type":"v4"},
				"2403.7000.8000.c00..9b": {"ip_addr":"2403:7000:8000:c00::9b","netmask":"ffff:ffff:ffff:ffff::","gateway":"2403:7000:8000:c00::1","location":"AKLCITY","type":"v6"}
			}
		}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListAllocatedIPs(context.Background())
	if err != nil {
		t.Fatalf("ListAllocatedIPs: %v", err)
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d", len(got.Return))
	}
	v4, ok := got.Return["203.0.113.10"]
	if !ok || v4.IPAddr != "203.0.113.10" || v4.Type != "v4" {
		t.Errorf("v4 entry = %+v", v4)
	}
	v6, ok := got.Return["2403.7000.8000.c00..9b"]
	if !ok || v6.IPAddr != "2403:7000:8000:c00::9b" {
		t.Errorf("v6 entry IPAddr = %q (key form is dotted; value preserves the literal)", v6.IPAddr)
	}
}

func TestListStatisticTypes_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/list_statistic_types.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != "ch-test" {
			t.Errorf("server_name = %q (note: not 'name')", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":["cpu","mem","disk"]}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListStatisticTypes(context.Background(), ListStatisticTypesOptions{ServerName: "ch-test"})
	if err != nil {
		t.Fatalf("ListStatisticTypes: %v", err)
	}
	if len(got.Return) != 3 {
		t.Errorf("Return = %v", got.Return)
	}
}

func TestGetStatistics_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/get_statistics.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != "ch-test" {
			t.Errorf("server_name = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"cpu":[1.0,2.0,3.0]}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).GetStatistics(context.Background(), GetStatisticsOptions{ServerName: "ch-test"})
	if err != nil {
		t.Fatalf("GetStatistics: %v", err)
	}
	if got.Return["cpu"] == nil {
		t.Errorf("Return['cpu'] missing; got %v", got.Return)
	}
}
