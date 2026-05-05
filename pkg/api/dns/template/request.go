package template

type (
	// GetRequest fetches one template's metadata + SOA defaults.
	//
	// TemplateID is a string because the API accepts and returns it
	// that way (e.g. "0", "855") and template_id "0" is the real
	// "Manual DNS Settings" platform-shared template — typing it as
	// int and rejecting the zero value would block legitimate calls.
	GetRequest struct {
		TemplateID string
	}

	// ListRecordsRequest lists every record under a template.
	ListRecordsRequest struct {
		TemplateID string
	}

	// SearchTemplatesRequest fuzzy-matches templates by name.
	// Offset / Limit are optional pagination knobs; both default to
	// the API default when zero.
	SearchTemplatesRequest struct {
		TemplateName string
		Offset       int
		Limit        int
	}

	// CreateTemplateRequest registers a new template. Name is the
	// only strictly-required field; the rest configure the SOA
	// defaults applied to every domain linked to the template.
	// Refresh / Retry / Expire / Min are TTL values in seconds.
	CreateTemplateRequest struct {
		Name       string
		Nameserver string
		Email      string
		Refresh    int
		Retry      int
		Expire     int
		Min        int
	}

	// CloneTemplateRequest duplicates an existing template under
	// a new name. The clone inherits records and SOA defaults.
	CloneTemplateRequest struct {
		TemplateID      string
		NewTemplateName string
	}

	// UpdateTemplateRequest renames an existing template. The API
	// only documents `new_name` as updatable here; for SOA-default
	// edits, recreate or use UpdateTemplateDNS to rebuild zones.
	UpdateTemplateRequest struct {
		TemplateID string
		NewName    string
	}

	// DeleteTemplateRequest removes a template. The API rejects the
	// call if any domain is still linked to it.
	DeleteTemplateRequest struct {
		TemplateID string
	}

	// AddRecordRequest adds a record to a template. Type is one of
	// the documented enum: A, AAAA, NS, MX, PTR, SRV, TXT, CNAME.
	// Priority is only meaningful for MX/SRV; pass 0 otherwise.
	AddRecordRequest struct {
		TemplateID string
		Type       string
		Name       string
		Content    string
		Priority   int
	}

	// UpdateRecordRequest replaces a record in-place. All fields
	// are required.
	UpdateRecordRequest struct {
		TemplateID string
		RecordID   int
		Type       string
		Name       string
		Content    string
		Priority   int
	}

	// DeleteRecordRequest removes one record from a template.
	DeleteRecordRequest struct {
		TemplateID string
		RecordID   int
	}

	// UpdateDomainRequest points a domain at a different template
	// (or unlinks it by passing TemplateID="").
	UpdateDomainRequest struct {
		Domain     string
		TemplateID string
	}

	// UpdateDomainDNSRequest forces a rebuild of one domain's zone
	// from its currently-linked template. Returns a scheduler job.
	UpdateDomainDNSRequest struct {
		Domain string
	}

	// UpdateTemplateDNSRequest forces a rebuild of every domain
	// linked to the given template. Returns a scheduler job.
	UpdateTemplateDNSRequest struct {
		TemplateID string
	}
)
