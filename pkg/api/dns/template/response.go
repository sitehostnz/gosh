package template

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// DomainTemplate is a single DNS domain template entry as
	// returned by ListTemplates / SearchTemplates.
	DomainTemplate struct {
		ClientID     string `json:"client_id"`
		TemplateID   string `json:"template_id"`
		TemplateName string `json:"template_name"`
		DomainCount  string `json:"domain_count"`
	}

	// TemplateDetails carries the full metadata + SOA defaults for
	// a single template, returned by GetTemplate. Fields are
	// strings to match the API's stringly-typed numerics.
	//
	//nolint:revive // name kept verbose for grep-from-API parity
	TemplateDetails struct {
		TemplateID   string `json:"template_id"`
		ClientID     string `json:"client_id"`
		TemplateName string `json:"template_name"`
		Nameserver   string `json:"nameserver"`
		Email        string `json:"email"`
		Refresh      string `json:"refresh"`
		Retry        string `json:"retry"`
		Expire       string `json:"expire"`
		Min          string `json:"min"`
		DomainCount  string `json:"domain_count"`

		// Both were being dropped: the API sends them and no field
		// received them. DateAdded can be "0000-00-00 00:00:00" on the
		// shared templates, which is MySQL's zero date rather than a
		// parseable time, so it is a string here.
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}

	// SearchResult is a single hit from SearchTemplates. Note the
	// shape differs from ListTemplates / GetTemplate: no ID field,
	// SOA defaults inline.
	SearchResult struct {
		ClientID     string `json:"client_id"`
		TemplateName string `json:"template_name"`
		Nameserver   string `json:"nameserver"`
		Email        string `json:"email"`
		Refresh      string `json:"refresh"`
		Retry        string `json:"retry"`
		Expire       string `json:"expire"`
		Min          string `json:"min"`
	}

	// Record is a single DNS record under a template.
	Record struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		Content    string `json:"content"`
		Priority   string `json:"prio"`
		ChangeDate string `json:"change_date"`
	}

	// ListResponse — list_templates.
	ListResponse struct {
		Return []DomainTemplate `json:"return"`
		models.APIResponse
	}

	// GetResponse — get_template returns an array, but in practice
	// always with a single element.
	GetResponse struct {
		Return []TemplateDetails `json:"return"`
		models.APIResponse
	}

	// ListRecordsResponse — list_records.
	ListRecordsResponse struct {
		Return []Record `json:"return"`
		models.APIResponse
	}

	// SearchTemplatesResponse — search_templates.
	SearchTemplatesResponse struct {
		Return []SearchResult `json:"return"`
		models.APIResponse
	}

	// CreateTemplateResponse — create_template returns the new
	// template's id (capitalised key per the API).
	CreateTemplateResponse struct {
		Return struct {
			TemplateID string `json:"TemplateID"`
		} `json:"return"`
		models.APIResponse
	}

	// CloneTemplateResponse — clone_template returns the new
	// template's id (lowercase key per the API).
	CloneTemplateResponse struct {
		Return struct {
			TemplateID string `json:"template_id"`
		} `json:"return"`
		models.APIResponse
	}

	// UpdateDomainResponse — update_domain returns a bare bool in
	// `return` plus the standard {msg, status} envelope.
	UpdateDomainResponse struct {
		Return bool `json:"return"`
		models.APIResponse
	}

	// JobResponse — shared shape for the async DNS-rebuild
	// endpoints (UpdateDomainDNS, UpdateTemplateDNS).
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
