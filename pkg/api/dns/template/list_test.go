package template_test

import (
	"context"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
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
}
