package securitygroups

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

// TestList_DecodesARecordedResponse covers an endpoint that had never
// decoded.
//
// servers was declared []string while the API sends
// [{"name":..., "label":...}], so every call failed with "cannot
// unmarshal object into Go struct field .return.data.servers of type
// string". GetResponse had the shape right the whole time, which is how
// the mistake survived: the two were written from different responses
// and only one of them was ever run.
func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(filepath.Join("testdata", "list_all.json"))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	c, err := api.New("k", "1", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := New(c).List(context.Background(), ListAllRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(got.Return.Data) == 0 {
		t.Fatal("no groups decoded; the fixture has one")
	}
	group := got.Return.Data[0]
	if group.Name == "" {
		t.Error("Name is empty; the name is what every other endpoint takes")
	}

	// The attached servers are the point of the listing, and the field
	// that never decoded.
	if len(group.Servers) == 0 {
		t.Fatal("Servers is empty; the fixture attaches one")
	}
	if group.Servers[0].Name == "" {
		t.Error("Servers[0].Name is empty; an attached server carries a name and a label, not a bare string")
	}
	if got.Return.TotalItems == 0 {
		t.Error("TotalItems = 0; the pagination envelope did not decode")
	}
}
