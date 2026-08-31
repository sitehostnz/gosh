package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// TestListProducts_DecodesWireQuirks covers the two response shapes that
// break a naive struct: attributes arriving as an empty array instead of
// an object, and partition sizes arriving as numbers instead of strings.
// Both were observed in a single live response.
func TestListProducts_DecodesWireQuirks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("location"); got != "AKLNCT" {
			t.Errorf("location = %q, want AKLNCT", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":[
			{"code":"LHPVS1","type":"HPVS","name":"Linux HPVS - 1 Core","price":"40","description":"",
			 "attributes":{"cores":1,"ram":2,"bandwidth":1024,"disk":50,
			   "partitions":[{"name":"scsi0","type":"ssd","size":"50"},{"name":"scsi1","type":"ssd","size":1}]}},
			{"code":"LSVSP1","type":"SVS","name":"Linux 1.5GB","price":"30","description":"",
			 "attributes":{"cores":1,"ram":1.5,"bandwidth":100,"disk":15,"partitions":[]}},
			{"code":"SOMETHING","type":"LICENS","name":"A licence","price":"10","description":"",
			 "attributes":[]},
			{"code":"CLDCON1","type":"CLDCON","name":"Cloud Container - 1 Core","price":"20","description":"",
			 "attributes":{"cores":1,"ram":1,"bandwidth":50,"containers":5,"images":2}}
		]}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	got, err := New(c).ListProducts(context.Background(), ListProductsOptions{Location: "AKLNCT"})
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(got.Return) != 4 {
		t.Fatalf("len(Return) = %d, want 4", len(got.Return))
	}

	// An unquoted partition size must not fail the whole decode.
	hpvs := got.Return[0]
	if n := len(hpvs.Attributes.Partitions); n != 2 {
		t.Fatalf("HPVS partitions = %d, want 2", n)
	}
	if s := hpvs.Attributes.Partitions[0].Size.String(); s != "50" {
		t.Errorf("quoted size = %q, want 50", s)
	}
	if s := hpvs.Attributes.Partitions[1].Size.String(); s != "1" {
		t.Errorf("unquoted size = %q, want 1", s)
	}
	if p := hpvs.Price.String(); p != "40" {
		t.Errorf("Price = %q, want 40", p)
	}

	// RAM is not always whole.
	if r := got.Return[1].Attributes.RAM; r != 1.5 {
		t.Errorf("LSVSP1 RAM = %v, want 1.5", r)
	}

	// attributes: [] must decode to a zero value, not an error.
	if lic := got.Return[2]; lic.Attributes.Cores != 0 || len(lic.Attributes.Extra) != 0 {
		t.Errorf("empty-array attributes decoded to %+v, want zero value", lic.Attributes)
	}

	// Attributes with no typed field are kept rather than dropped.
	cld := got.Return[3]
	if _, ok := cld.Attributes.Extra["containers"]; !ok {
		t.Errorf("Extra = %v, want it to carry \"containers\"", cld.Attributes.Extra)
	}
	if cld.Attributes.Cores != 1 {
		t.Errorf("Cores = %d, want 1 — typed fields must still decode alongside Extra", cld.Attributes.Cores)
	}
}

// TestListProducts_FiltersOnWire checks both filters reach the API in the
// array form it expects, without duplicating values.
func TestListProducts_FiltersOnWire(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q["filters[type][]"]; len(got) != 2 || got[0] != "HPVS" || got[1] != "SVS" {
			t.Errorf("filters[type][] = %v, want [HPVS SVS]", got)
		}
		if got := q["filters[code][]"]; len(got) != 1 || got[0] != "LHPVS1" {
			t.Errorf("filters[code][] = %v, want [LHPVS1]", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful","return":[]}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	if _, err := New(c).ListProducts(context.Background(), ListProductsOptions{
		Location: "AKLNCT",
		Types:    []string{"HPVS", "SVS"},
		Codes:    []string{"LHPVS1"},
	}); err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
}

// TestListProducts_RequiresLocation checks the pairing is enforced before
// a request goes out; products are scoped to a location's product group,
// so there is no "list everything" call.
func TestListProducts_RequiresLocation(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("no request should be made without a location")
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	if _, err := New(c).ListProducts(context.Background(), ListProductsOptions{}); err == nil {
		t.Fatal("ListProducts: expected an error when Location is empty")
	}
}
