package template_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/dns/template"
)

func TestList_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_templates.json")

	got, err := template.New(ex.Client).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, template.ListResponse{})

	// The request, not just the response. apitest.Serve answers the
	// fixture whatever is asked for, so without these the method and
	// the endpoint are unpinned: both were mutated — POST to
	// list_TYPO.json — and this test still passed.
	if ex.Request.Method != "GET" {
		t.Errorf("method = %s, want GET", ex.Request.Method)
	}
	if ex.Request.URL.Path != "/dns/domain_templates/list_templates.json" {
		t.Errorf("path = %s, want /dns/domain_templates/list_templates.json", ex.Request.URL.Path)
	}
	q := ex.Request.URL.Query()
	if q.Get("apikey") == "" {
		t.Error("apikey is absent from the request")
	}
	if q.Get("client_id") == "" {
		t.Error("client_id is absent from the request")
	}

	if len(got.Return) == 0 {
		t.Fatal("no templates decoded; the fixture has rows")
	}
	for i, tpl := range got.Return {
		if tpl.TemplateID == "" {
			t.Errorf("Return[%d].TemplateID is empty", i)
		}
	}
}

// TestGet_ReturnsAnArray pins a shape that reads like a mistake and is
// not: get_template answers with a list holding one element, rather
// than the object the name implies.
func TestGet_ReturnsAnArray(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "get_template.json")

	got, err := template.New(ex.Client).Get(context.Background(),
		template.GetRequest{TemplateID: "0"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, template.GetResponse{})

	if len(got.Return) != 1 {
		t.Fatalf("Return has %d element(s), want 1 — this endpoint answers with a list", len(got.Return))
	}
	if got.Return[0].Nameserver == "" {
		t.Error("Nameserver is empty; the SOA fields are what distinguishes Get from List")
	}
}

// TestListRecords_RejectsAnUnknownTemplate records that this endpoint
// reports absence through an error, unlike dns.GetZone. Note also that
// template id "0" is a real template rather than a null id, so it is
// not a usable probe for absence.
func TestListRecords_RejectsAnUnknownTemplate(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_records-unknown.json")

	_, err := template.New(ex.Client).ListRecords(context.Background(),
		template.ListRecordsRequest{TemplateID: "99999999"})
	if err == nil {
		t.Fatal("ListRecords: expected an error for a template that does not exist")
	}
	// The API's own message has to reach the caller's error, or a
	// consumer cannot tell a not-found from a permissions failure or a
	// throttle. Asserting only err != nil left nothing in this package
	// checking that a rejection's message survives.
	if !strings.Contains(err.Error(), "doesnt exist") {
		t.Errorf("ListRecords: error is %q, want the API's message to survive into it", err)
	}
	apitest.AssertDecodesFully(t, ex.Body, template.ListRecordsResponse{})
}

// TestList_SharedTemplatesCarryClientIDZero pins the sentinel the doc
// tells consumers to filter on.
//
// It is hand-written rather than fixture-backed because a fixture
// cannot express it: internal/recorder/scrub.go collapses any
// digits-only string to "1", so a real "0" is unrepresentable in
// committed testdata and testdata/list_templates.json accordingly
// carries "client_id": "1" on every row. The advice in the doc — filter
// on ClientID to separate SiteHost's shared templates from the
// account's own — was therefore unverifiable from anything in the tree.
func TestList_SharedTemplatesCarryClientIDZero(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status": true,
			"msg": "Successful",
			"return": [
				{"template_id": "0", "template_name": "Manual DNS Settings", "client_id": "0", "domain_count": "412"},
				{"template_id": "9931", "template_name": "our own template", "client_id": "979387", "domain_count": "3"}
			]
		}`)
	}))
	defer srv.Close()

	c, err := api.New("k", "979387", api.SetBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	got, err := template.New(c).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Return) != 2 {
		t.Fatalf("len(Return) = %d, want 2", len(got.Return))
	}

	// The mapping first: a swapped json tag would make the filter below
	// pass for the wrong reason.
	if got.Return[0].ClientID != "0" {
		t.Errorf("Return[0].ClientID = %q, want \"0\"", got.Return[0].ClientID)
	}
	if got.Return[0].TemplateName != "Manual DNS Settings" {
		t.Errorf("Return[0].TemplateName = %q, want the shared template's name", got.Return[0].TemplateName)
	}
	if got.Return[1].ClientID != "979387" {
		t.Errorf("Return[1].ClientID = %q, want the account id", got.Return[1].ClientID)
	}

	// Then the advice the doc gives.
	var own []string
	for _, tpl := range got.Return {
		if tpl.ClientID != "0" {
			own = append(own, tpl.TemplateName)
		}
	}
	if len(own) != 1 || own[0] != "our own template" {
		t.Errorf("filtering on ClientID != \"0\" gave %v, want just the account's own", own)
	}
}
