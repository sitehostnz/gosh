package mail

type (
	// ServerOptions identifies a mail server (the API's
	// server_name parameter). Embedded into operation-specific
	// option structs. Leave ServerName empty to inherit from a
	// NewForServer-captured default; otherwise it is required.
	ServerOptions struct {
		ServerName string `url:"server_name"`
	}

	// GetServerInfoOptions identifies the mail server to fetch
	// info about.
	GetServerInfoOptions struct {
		ServerOptions
	}

	// ListDomainsOptions identifies the mail server whose domains
	// to list.
	ListDomainsOptions struct {
		ServerOptions
	}

	// GetAccountOptions identifies the mail server and the email
	// address whose account to fetch.
	GetAccountOptions struct {
		ServerOptions
		Email string `url:"email"`
	}

	// ListAccountsOptions identifies the mail server and domain
	// whose accounts to list. EmailAddr is an optional filter
	// restricting results to a specific address.
	ListAccountsOptions struct {
		ServerOptions
		Domain    string `url:"domain"`
		EmailAddr string `url:"filters[emailaddr],omitempty"`
	}

	// ListAllOptions identifies the mail server and domain whose
	// every-email-record listing to fetch (mailboxes + aliases +
	// forwarders, in a single union view).
	ListAllOptions struct {
		ServerOptions
		Domain string `url:"domain"`
	}

	// SearchAccountsOptions identifies the mail server to search
	// against. At least one of EmailAddr / Username / Active /
	// Quota should be set; the API rejects calls with no query[*]
	// filter.
	SearchAccountsOptions struct {
		ServerOptions
		EmailAddr string `url:"query[emailaddr],omitempty"`
		Username  string `url:"query[username],omitempty"`
		Active    string `url:"query[active],omitempty"`
		Quota     string `url:"query[quota],omitempty"`
		Offset    int    `url:"offsets[offset],omitempty"`
		Limit     int    `url:"offsets[limit],omitempty"`
	}

	// ListAliasesOptions identifies the mail server and domain
	// whose aliases to list. Source is an optional filter
	// restricting results to a specific source address.
	ListAliasesOptions struct {
		ServerOptions
		Domain string `url:"domain"`
		Source string `url:"filters[source],omitempty"`
	}

	// ListForwardsOptions identifies the mail server and domain
	// whose forwarders to list. Source is an optional filter
	// restricting results to a specific source address.
	ListForwardsOptions struct {
		ServerOptions
		Domain string `url:"domain"`
		Source string `url:"filters[source],omitempty"`
	}

	// SearchAliasesOptions identifies the mail server to search
	// against for aliases. At least one of Source or Destination
	// must be set; the API rejects calls with no query[*] filter.
	SearchAliasesOptions struct {
		ServerOptions
		Source      string `url:"query[source],omitempty"`
		Destination string `url:"query[destination],omitempty"`
		Offset      int    `url:"offsets[offset],omitempty"`
		Limit       int    `url:"offsets[limit],omitempty"`
	}

	// SearchForwardsOptions identifies the mail server to search
	// against for forwarders. At least one of Source or
	// Destination must be set.
	SearchForwardsOptions struct {
		ServerOptions
		Source      string `url:"query[source],omitempty"`
		Destination string `url:"query[destination],omitempty"`
		Offset      int    `url:"offsets[offset],omitempty"`
		Limit       int    `url:"offsets[limit],omitempty"`
	}

	// AccountParams holds the optional per-account settings shared
	// by AddAccount and UpdateAccount. For AddAccount the
	// Password field is required.
	AccountParams struct {
		Password          string `url:"params[password],omitempty"`
		Label             string `url:"params[label],omitempty"`
		Username          string `url:"params[username],omitempty"`
		Autoresponder     string `url:"params[autoresponder],omitempty"`
		AutoresponderText string `url:"params[autoresponder_text],omitempty"`
		SpamStrategy      string `url:"params[spam_strategy],omitempty"`
		Quota             string `url:"params[quota],omitempty"`
	}

	// AddAccountOptions describes a new mail account to create.
	// ServerName, Email, and Password are required; the rest are
	// optional. (AccountParams is embedded anonymously so its
	// url:"params[...]" tags are picked up by go-querystring.)
	AddAccountOptions struct {
		ServerOptions
		Email string `url:"email"`
		AccountParams
	}

	// UpdateAccountOptions describes an account update. All
	// AccountParams fields are optional; supplying none is a no-op.
	UpdateAccountOptions struct {
		ServerOptions
		Email string `url:"email"`
		AccountParams
	}

	// DeleteAccountOptions identifies the account to delete.
	DeleteAccountOptions struct {
		ServerOptions
		Email string `url:"email"`
	}

	// AddDomainOptions describes a domain to add to the mail server.
	//
	// Domain is validated against a TLD allowlist server-side.
	// RFC 2606 reserved test TLDs (.example, .test, .localhost,
	// .invalid) are rejected with "Please specify a valid domain
	// name." Use a domain on a real public-suffix TLD, or one of
	// SiteHost's wildcard test domains (sth.nz) for SDK testing.
	AddDomainOptions struct {
		ServerOptions
		Domain string `url:"domain"`
	}

	// DomainParams holds the optional per-domain settings for
	// UpdateDomain.
	DomainParams struct {
		Catchall string `url:"params[catchall],omitempty"`
		State    string `url:"params[state],omitempty"`
	}

	// UpdateDomainOptions describes a domain update.
	UpdateDomainOptions struct {
		ServerOptions
		Domain string `url:"domain"`
		DomainParams
	}

	// DeleteDomainOptions identifies the domain to remove.
	DeleteDomainOptions struct {
		ServerOptions
		Domain string `url:"domain"`
	}

	// AliasMapping is the source→destination pair used by add/
	// delete operations for both aliases and forwarders. Note the
	// delete operations require both fields — the API rejects
	// calls with only the source.
	AliasMapping struct {
		ServerOptions
		Source      string `url:"source"`
		Destination string `url:"destination"`
	}

	// AddAliasOptions adds an alias mapping.
	AddAliasOptions = AliasMapping

	// DeleteAliasOptions deletes an alias mapping.
	DeleteAliasOptions = AliasMapping

	// AddForwardOptions adds a forwarder mapping.
	AddForwardOptions = AliasMapping

	// DeleteForwardOptions deletes a forwarder mapping.
	DeleteForwardOptions = AliasMapping

	// AddAliasDomainOptions adds an alias-domain mapping (one
	// domain pointing at another for mail purposes).
	AddAliasDomainOptions struct {
		ServerOptions
		AliasDomain  string `url:"alias_domain"`
		ParentDomain string `url:"parent_domain"`
	}

	// DeleteAliasDomainOptions removes an alias-domain mapping.
	DeleteAliasDomainOptions struct {
		ServerOptions
		AliasDomain string `url:"alias_domain"`
	}
)
