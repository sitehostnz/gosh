package srs

type (
	// ListDomainsOptions represents optional filters for the list_domains
	// call. SortBy and SortDir control ordering; PageSize and PageNumber
	// control pagination. All fields are optional.
	ListDomainsOptions struct {
		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}

	// DomainOptions identifies a single domain — used by get_domain,
	// domain_inside_grace_period, get_domain_price,
	// can_transfer_domain, list_name_servers, and the lifecycle
	// operations (cancel_domain, lock_domain, unlock_domain).
	DomainOptions struct {
		Domain string `url:"domain"`
	}

	// DomainAvailableOptions checks whether a domain is registrable.
	DomainAvailableOptions struct {
		Domain string `url:"domain"`
	}

	// CreateDomainOptions describes a new .nz domain to register.
	// Domain and the four contact IDs are required.
	//
	// API parameter naming is inconsistent here:
	//   - registrant_contact (no _id, top-level)
	//   - params[AdminContact] (PascalCase, nested in params)
	//   - params[TechContact]  (note: "Tech", not "Technical")
	//   - params[BillingContact]
	// The Go fields use uniform names; tags reflect the wire shape.
	CreateDomainOptions struct {
		Domain             string `url:"domain"`
		Term               int    `url:"term,omitempty"`
		RegistrantContact  int    `url:"registrant_contact"`
		AdminContact       int    `url:"params[AdminContact]"`
		TechnicalContact   int    `url:"params[TechContact]"`
		BillingContact     int    `url:"params[BillingContact]"`
		Privacy            string `url:"privacy,omitempty"`
	}
)
