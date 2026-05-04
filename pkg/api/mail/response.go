package mail

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ServerInfo describes a mail server's connection details.
	ServerInfo struct {
		Hostname    string `json:"hostname"`
		WebmailURL  string `json:"webmail_url"`
		DateAdded   string `json:"date_added"`
		DateUpdated string `json:"date_updated"`
	}

	// GetServerInfoResponse represents the response from
	// get_server_info.
	GetServerInfoResponse struct {
		Return ServerInfo `json:"return"`
		models.APIResponse
	}

	// Domain is a single mail-domain entry from list_domains.
	// State is returned as a string ("1" enabled / "0" disabled).
	// ParentDomain is non-empty for alias domains pointing at
	// another domain; otherwise empty.
	// Accounts / Nicknames / Forwarders / TotalUsed are returned
	// as JSON numbers (so int in Go), distinct from the
	// string-typed counters elsewhere in the API.
	Domain struct {
		ClientID     string `json:"client_id"`
		Domain       string `json:"domain"`
		ParentDomain string `json:"parent_domain"`
		CatchAll     string `json:"catch_all"`
		State        string `json:"state"`
		Accounts     int    `json:"accounts"`
		Nicknames    int    `json:"nicknames"`
		Forwarders   int    `json:"forwarders"`
		TotalUsed    int    `json:"total_used"`
	}

	// ListDomainsResponse represents the response from
	// list_domains.
	ListDomainsResponse struct {
		Return []Domain `json:"return"`
		models.APIResponse
	}

	// EmailRecord is a single row from /mail/list_all.json — the
	// union of mailboxes, aliases, and forwarders. Type
	// distinguishes the kind:
	//
	//   0  mailbox      Username + EmailAddr + Label populated
	//   1  alias        EmailAddr + Destination populated
	//   2  forward      EmailAddr + Destination populated
	//
	// Fields not relevant to a row's Type deserialise to empty.
	EmailRecord struct {
		Type        int    `json:"type"`
		Username    string `json:"username,omitempty"`
		EmailAddr   string `json:"emailaddr"`
		Label       string `json:"label,omitempty"`
		Destination string `json:"destination,omitempty"`
	}

	// ListAllResponse is the return from /mail/list_all.json.
	ListAllResponse struct {
		Return []EmailRecord `json:"return"`
		models.APIResponse
	}

	// Account is a mail account record. Field availability varies
	// by the source endpoint:
	//
	//   - search_accounts populates the first 9 fields only.
	//   - list_accounts adds QuotaUsed, QuotaPercent, MessageCount,
	//     LastUpdated.
	//   - get_account additionally populates Key and DateAdded.
	//
	// Fields not populated by an endpoint deserialise to their
	// zero value (empty string for the string-typed fields). All
	// numeric / boolean values are returned as strings by the API
	// ("0", "1", "yes", "no") and represented as Go strings here
	// for fidelity.
	Account struct {
		ClientID          string `json:"client_id"`
		EmailAddr         string `json:"emailaddr"`
		Label             string `json:"label"`
		Username          string `json:"username"`
		Autoresponder     string `json:"autoresponder"`
		AutoresponderText string `json:"autoresponder_text"`
		Active            string `json:"active"`
		Quota             string `json:"quota"`
		SpamStrategy      string `json:"spam_strategy"`
		QuotaUsed         string `json:"quota_used,omitempty"`
		QuotaPercent      string `json:"quota_percent,omitempty"`
		MessageCount      string `json:"message_count,omitempty"`
		LastUpdated       string `json:"last_updated,omitempty"`
		Key               string `json:"key,omitempty"`
		DateAdded         string `json:"date_added,omitempty"`
	}

	// GetAccountResponse represents the response from get_account.
	GetAccountResponse struct {
		Return Account `json:"return"`
		models.APIResponse
	}

	// ListAccountsResponse represents the response from
	// list_accounts.
	ListAccountsResponse struct {
		Return []Account `json:"return"`
		models.APIResponse
	}

	// SearchAccountsResponse represents the response from
	// search_accounts.
	SearchAccountsResponse struct {
		Return []Account `json:"return"`
		models.APIResponse
	}

	// Alias is a mail alias mapping (one address redirecting to
	// another). Aliases are the same data shape as forwards
	// (source → destination); the type names exist to mirror the
	// API's separate endpoints.
	//
	// ClientID is populated by search_aliases but not by
	// list_aliases.
	Alias struct {
		ClientID    string `json:"client_id,omitempty"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}

	// ListAliasesResponse represents the response from
	// list_aliases.
	ListAliasesResponse struct {
		Return []Alias `json:"return"`
		models.APIResponse
	}

	// SearchAliasesResponse represents the response from
	// search_aliases.
	SearchAliasesResponse struct {
		Return []Alias `json:"return"`
		models.APIResponse
	}

	// Forward is a mail forwarder mapping (a source address
	// forwarding to a destination, typically external).
	//
	// ClientID is populated by search_forwards but not by
	// list_forwards.
	Forward struct {
		ClientID    string `json:"client_id,omitempty"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}

	// ListForwardsResponse represents the response from
	// list_forwards.
	ListForwardsResponse struct {
		Return []Forward `json:"return"`
		models.APIResponse
	}

	// SearchForwardsResponse represents the response from
	// search_forwards.
	SearchForwardsResponse struct {
		Return []Forward `json:"return"`
		models.APIResponse
	}

	// JobResponse is the response shape for the mail write
	// operations that queue an asynchronous job: AddAccount,
	// UpdateAccount, DeleteAccount, AddAlias, AddForward. The
	// returned job ID is for tracking only.
	//
	// The other write operations (AddDomain, UpdateDomain,
	// DeleteDomain, DeleteAlias, DeleteForward, AddAliasDomain,
	// DeleteAliasDomain) take effect synchronously and return only
	// models.APIResponse.
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
