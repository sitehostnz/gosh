package template

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api"
)

func formBody(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	v, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return v
}

func TestGet_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/domain_templates/get_template.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("template_id") != "168" {
			t.Errorf("template_id = %q", r.URL.Query().Get("template_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful.",
			"return":[{"template_id":"168","client_id":"1","template_name":"New Template","nameserver":"ns1.sitehost.co.nz","email":"example@example.com","refresh":"3600","retry":"3600","expire":"3600","min":"3600","domain_count":"0"}]
		}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).Get(context.Background(), GetRequest{TemplateID: "168"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Return) != 1 || got.Return[0].TemplateName != "New Template" {
		t.Errorf("Return = %+v", got.Return)
	}
}

func TestGet_RequiresTemplateID(t *testing.T) {
	t.Parallel()
	c, _ := api.New("k", "1", api.SetBaseURL("http://example.invalid"))
	if _, err := New(c).Get(context.Background(), GetRequest{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestListRecords_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/domain_templates/list_records.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"status":true,"msg":"Successful.",
			"return":[
				{"id":"1459","name":"@","type":"NS","content":"ns1.sitehost.co.nz.","prio":"0","change_date":"1461119775"},
				{"id":"1462","name":"subdomain","type":"A","content":"1.1.1.1","prio":"0","change_date":"1461119775"}
			]
		}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).ListRecords(context.Background(), ListRecordsRequest{TemplateID: "12"})
	if err != nil {
		t.Fatalf("ListRecords: %v", err)
	}
	if len(got.Return) != 2 || got.Return[1].Type != "A" {
		t.Errorf("Return = %+v", got.Return)
	}
}

func TestSearchTemplates_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/domain_templates/search_templates.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("query[template_name]") != "New" {
			t.Errorf("query = %v", v)
		}
		if v.Get("limitArr[offset]") != "5" || v.Get("limitArr[limit]") != "10" {
			t.Errorf("paging = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":[{"client_id":"1","template_name":"New Example Template","nameserver":"ns1.sitehost.co.nz","email":"soa@sitehost.co.nz","refresh":"10800","retry":"3600","expire":"604800","min":"3600"}]}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).SearchTemplates(context.Background(), SearchTemplatesRequest{
		TemplateName: "New", Offset: 5, Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchTemplates: %v", err)
	}
	if len(got.Return) != 1 || got.Return[0].TemplateName != "New Example Template" {
		t.Errorf("Return = %+v", got.Return)
	}
}

func TestCreateTemplate_AllOptional(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("name") != "T1" {
			t.Errorf("name = %q", v.Get("name"))
		}
		if v.Get("params[Nameserver]") != "ns1.sitehost.co.nz" {
			t.Errorf("ns = %q", v.Get("params[Nameserver]"))
		}
		if v.Get("params[Refresh]") != "3600" {
			t.Errorf("refresh = %q", v.Get("params[Refresh]"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"TemplateID":"168"}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).CreateTemplate(context.Background(), CreateTemplateRequest{
		Name: "T1", Nameserver: "ns1.sitehost.co.nz", Refresh: 3600,
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if got.Return.TemplateID != "168" {
		t.Errorf("TemplateID = %q", got.Return.TemplateID)
	}
}

func TestCreateTemplate_OmitsZeroFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		for _, k := range []string{"params[Email]", "params[Refresh]", "params[Retry]", "params[Expire]", "params[Min]", "params[Nameserver]"} {
			if _, ok := v[k]; ok {
				t.Errorf("%s should be omitted, got %v", k, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"TemplateID":"1"}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).CreateTemplate(context.Background(), CreateTemplateRequest{Name: "Bare"}); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
}

func TestCloneTemplate_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("template_id") != "1" || v.Get("new_template_name") != "Clone" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"template_id":"167"}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).CloneTemplate(context.Background(), CloneTemplateRequest{
		TemplateID: "1", NewTemplateName: "Clone",
	})
	if err != nil {
		t.Fatalf("CloneTemplate: %v", err)
	}
	if got.Return.TemplateID != "167" {
		t.Errorf("TemplateID = %q", got.Return.TemplateID)
	}
}

func TestUpdateTemplate_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("template_id") != "1" || v.Get("new_name") != "Renamed" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateTemplate(context.Background(), UpdateTemplateRequest{
		TemplateID: "1", NewName: "Renamed",
	}); err != nil {
		t.Fatalf("UpdateTemplate: %v", err)
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("template_id") != "15" {
			t.Errorf("template_id = %q", v.Get("template_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).DeleteTemplate(context.Background(), DeleteTemplateRequest{TemplateID: "15"}); err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}
}

func TestAddRecord_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("type") != "A" || v.Get("name") != "subdomain" || v.Get("content") != "1.1.1.1" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).AddRecord(context.Background(), AddRecordRequest{
		TemplateID: "1", Type: "A", Name: "subdomain", Content: "1.1.1.1",
	}); err != nil {
		t.Fatalf("AddRecord: %v", err)
	}
}

func TestUpdateRecord_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("record_id") != "12" || v.Get("content") != "192.168.1.10" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).UpdateRecord(context.Background(), UpdateRecordRequest{
		TemplateID: "1", RecordID: 12, Type: "A", Name: "x", Content: "192.168.1.10",
	}); err != nil {
		t.Fatalf("UpdateRecord: %v", err)
	}
}

func TestDeleteRecord_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("template_id") != "36" || v.Get("record_id") != "12" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful."}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	if _, err := New(c).DeleteRecord(context.Background(), DeleteRecordRequest{
		TemplateID: "36", RecordID: 12,
	}); err != nil {
		t.Fatalf("DeleteRecord: %v", err)
	}
}

func TestUpdateDomain_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := formBody(t, r)
		if v.Get("domain") != "example.com" || v.Get("params[template_id]") != "23" {
			t.Errorf("body = %v", v)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"Successful.","return":true}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpdateDomain(context.Background(), UpdateDomainRequest{
		Domain: "example.com", TemplateID: "23",
	})
	if err != nil {
		t.Fatalf("UpdateDomain: %v", err)
	}
	if !got.Return {
		t.Errorf("Return = %v", got.Return)
	}
}

func TestUpdateDomainDNS_ReturnsJob(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/domain_templates/update_domain_dns.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"job":{"type":"scheduler","id":7319498}}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpdateDomainDNS(context.Background(), UpdateDomainDNSRequest{Domain: "x.com"})
	if err != nil {
		t.Fatalf("UpdateDomainDNS: %v", err)
	}
	if got.Return.ID != 7319498 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}

func TestUpdateTemplateDNS_ReturnsJob(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dns/domain_templates/update_template_dns.json" {
			t.Errorf("path = %q", r.URL.Path)
		}
		v := formBody(t, r)
		if v.Get("template_id") != "9" {
			t.Errorf("template_id = %q", v.Get("template_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"msg":"OK","return":{"job":{"type":"scheduler","id":7319505}}}`)
	}))
	defer srv.Close()
	c, _ := api.New("k", "1", api.SetBaseURL(srv.URL))
	got, err := New(c).UpdateTemplateDNS(context.Background(), UpdateTemplateDNSRequest{TemplateID: "9"})
	if err != nil {
		t.Fatalf("UpdateTemplateDNS: %v", err)
	}
	if got.Return.ID != 7319505 {
		t.Errorf("Job.ID = %d", got.Return.ID)
	}
}
