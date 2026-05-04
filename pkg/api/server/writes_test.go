package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

const testIP = "192.0.2.10"

func mockClient(t *testing.T, h http.HandlerFunc) (*api.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		srv.Close()
		t.Fatalf("api.New: %v", err)
	}
	return c, srv.Close
}

func TestAddIP_Success(t *testing.T) {
	t.Parallel()
	c, done := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/server/add_ip.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = r.ParseForm()
		if got := r.Form.Get("param"); got != testIP {
			t.Errorf("param = %q (note: add_ip uses 'param', not 'address')", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"},"ip_addr":"192.0.2.10"}}`)
	})
	defer done()

	got, err := New(c).AddIP(context.Background(), AddIPOptions{Name: "ch-foo", IP: testIP})
	if err != nil {
		t.Fatalf("AddIP: %v", err)
	}
	if got.Return.IPAddr != testIP {
		t.Errorf("IPAddr = %q", got.Return.IPAddr)
	}
}

func TestRemoveIP_Success(t *testing.T) {
	t.Parallel()
	c, done := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.Form.Get("address"); got != testIP {
			t.Errorf("address = %q (note: remove_ip uses 'address', not 'param')", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"},"ip_addr":"192.0.2.10"}}`)
	})
	defer done()

	if _, err := New(c).RemoveIP(context.Background(), RemoveIPOptions{Name: "ch-foo", IP: testIP}); err != nil {
		t.Fatalf("RemoveIP: %v", err)
	}
}

func TestSetPrimaryIP_Success(t *testing.T) {
	t.Parallel()
	c, done := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.Form.Get("address"); got != testIP {
			t.Errorf("address = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"ip_addr":"192.0.2.10"}}`)
	})
	defer done()

	got, err := New(c).SetPrimaryIP(context.Background(), SetPrimaryIPOptions{Name: "ch-foo", IP: testIP})
	if err != nil {
		t.Fatalf("SetPrimaryIP: %v", err)
	}
	if got.Return.IPAddr != testIP {
		t.Errorf("IPAddr = %q", got.Return.IPAddr)
	}
}

func TestChangeState_Success(t *testing.T) {
	t.Parallel()
	c, done := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.Form.Get("state"); got != StatePowerOn {
			t.Errorf("state = %q, want power_on", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":1,"type":"scheduler"}}}`)
	})
	defer done()

	if _, err := New(c).ChangeState(context.Background(), ChangeStateOptions{Name: "ch-foo", State: StatePowerOn}); err != nil {
		t.Fatalf("ChangeState: %v", err)
	}
}

func TestCanProvision_Success(t *testing.T) {
	t.Parallel()
	c, done := mockClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.Form.Get("product"); got != "XENLIT" {
			t.Errorf("product = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	})
	defer done()

	got, err := New(c).CanProvision(context.Background(), CanProvisionOptions{
		Product: "XENLIT", Location: "AKLCITY", Distro: "ubuntu",
	})
	if err != nil {
		t.Fatalf("CanProvision: %v", err)
	}
	if !got.Status {
		t.Errorf("Status = false")
	}
}

func TestServerWrites_RequiredFields(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1")
	if _, err := New(c).AddIP(context.Background(), AddIPOptions{IP: "1.2.3.4"}); err == nil ||
		!strings.Contains(err.Error(), "Name is required") {
		t.Errorf("AddIP missing Name: %v", err)
	}
	if _, err := New(c).RemoveIP(context.Background(), RemoveIPOptions{Name: "n"}); err == nil ||
		!strings.Contains(err.Error(), "IP is required") {
		t.Errorf("RemoveIP missing IP: %v", err)
	}
	if _, err := New(c).ChangeState(context.Background(), ChangeStateOptions{Name: "n"}); err == nil ||
		!strings.Contains(err.Error(), "State is required") {
		t.Errorf("ChangeState missing State: %v", err)
	}
	if _, err := New(c).CanProvision(context.Background(), CanProvisionOptions{Location: "L", Distro: "d"}); err == nil ||
		!strings.Contains(err.Error(), "Product is required") {
		t.Errorf("CanProvision missing Product: %v", err)
	}
}
