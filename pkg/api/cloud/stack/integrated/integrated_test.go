package integrated

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// testServer is the server name every test here addresses.
const testServer = "ch-example"

// TestList_Success pins the response shape to what the API actually
// sends.
//
// The body here is a live response, not an assumed one: a bare array of
// stack names under "return", with no data wrapper and no pagination.
// Every other cloud listing uses {"return":{"data":[...]}}, so a
// fixture written from the family's habits rather than from the wire
// would decode to nothing and the test would still pass.
func TestList_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/integrated/list_all.json" {
			t.Errorf("path = %q, want /cloud/stack/integrated/list_all.json", r.URL.Path)
		}
		if got := r.URL.Query().Get("server_name"); got != testServer {
			t.Errorf("server_name = %q, want ch-example", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":["mariadb1108","mysql57"]}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).List(context.Background(), ListRequest{ServerName: testServer})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}
	if got.Return[0] != "mariadb1108" || got.Return[1] != "mysql57" {
		t.Errorf("Return = %v, want [mariadb1108 mysql57]", got.Return)
	}
}

// TestList_EmptyIsNotAnError covers a server with no database stack.
//
// That is the ordinary state of a freshly provisioned container server,
// and it is the case the caller has to distinguish: nowhere to put a
// database yet, rather than a failed call.
func TestList_EmptyIsNotAnError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":[]}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).List(context.Background(), ListRequest{ServerName: testServer})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Return) != 0 {
		t.Errorf("len(Return) = %d, want 0", len(got.Return))
	}
}

// TestList_RequiresServerName checks the guard fires before a request.
func TestList_RequiresServerName(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("no request should be made without a server name")
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).List(context.Background(), ListRequest{}); err == nil {
		t.Fatal("List: expected an error when ServerName is empty")
	}
}

// TestAdd_SendsServerAndName pins the parameter names.
//
// "server", not "server_name" — the two differ across the cloud
// endpoints, and net.Encode emits only the keys it is given, so a
// mismatch here would send nothing and the call would fail for a
// reason that looks nothing like a typo.
func TestAdd_SendsServerAndName(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cloud/stack/integrated/add.json" {
			t.Errorf("path = %q, want /cloud/stack/integrated/add.json", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.PostForm.Get("server"); got != testServer {
			t.Errorf("server = %q, want ch-example", got)
		}
		if got := r.PostForm.Get("name"); got != StackMariaDB1108 {
			t.Errorf("name = %q, want %q", got, StackMariaDB1108)
		}
		// The endpoint takes neither of these, and sending them would
		// mean this wrapper had drifted towards cloud/stack/add.json.
		if r.PostForm.Has("docker_compose") || r.PostForm.Has("label") {
			t.Errorf("unexpected compose/label parameters: %v", r.PostForm)
		}
		w.Header().Set("Content-Type", "application/json")
		// Captured from a live 1.5 response. Version 1.0 returns a
		// bare "job_id" string instead, and a fixture written from
		// that source decodes this into an empty job.
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job":{"id":31160001,"type":"scheduler"}}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Add(context.Background(), AddRequest{
		ServerName: testServer, Name: StackMariaDB1108,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if got.Return.ID != 31160001 {
		t.Errorf("job id = %d, want 31160001", got.Return.ID)
	}
	if got.Return.Type != "scheduler" {
		t.Errorf("job type = %q, want scheduler", got.Return.Type)
	}
}

// TestAdd_RejectsThe10ResponseShape pins the version difference that
// already caused a bug here.
//
// Version 1.0 of this endpoint returns {"job_id":"123"}; 1.5 returns
// {"job":{"id":123,"type":"scheduler"}}. This wrapper was first written
// from the 1.0 source, so it decoded a live 1.5 response into an empty
// job — and an empty job is worse than an error, because job.Get reads
// ID == 0 as "nothing to wait for" and the caller carries on as though
// a stack it never waited for had been built.
//
// So a 1.0-shaped body must not silently yield a usable job.
func TestAdd_RejectsThe10ResponseShape(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Asserted here too, not only in the test above: a test that
		// sends a parameter and never looks for it would pass if the
		// client stopped sending it, which reads as coverage and is
		// not.
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.PostForm.Get("server"); got != testServer {
			t.Errorf("server = %q, want ch-example", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":{"job_id":"31160001"}}`)
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Add(context.Background(), AddRequest{
		ServerName: testServer, Name: StackMySQL84,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	// The point of the test: this is what the old shape produces, and a
	// caller must not be able to mistake it for a job to poll.
	if got.Return.ID != 0 {
		t.Fatalf("job id = %d; a 1.0-shaped body should not decode into a job", got.Return.ID)
	}
}

// TestAdd_RequiresNameAndServer checks both guards fire locally.
func TestAdd_RequiresNameAndServer(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("no request should be made with missing required fields")
	}))
	defer srv.Close()

	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	cl := New(c)

	if _, err := cl.Add(context.Background(), AddRequest{Name: StackMySQL57}); err == nil {
		t.Error("Add: expected an error when ServerName is empty")
	}
	if _, err := cl.Add(context.Background(), AddRequest{ServerName: testServer}); err == nil {
		t.Error("Add: expected an error when Name is empty")
	}
}
