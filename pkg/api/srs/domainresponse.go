package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// Domain is a single registered-domain entry from list_domains.
	//
	// String-typed numeric / boolean fields (Locked, Private, Pending,
	// Premium, AutoRenewTerm, AutoRenewDaysRemaining) reflect the API's
	// actual response shape — values arrive as strings ("0", "1",
	// "12") rather than typed numbers / bools. Consumers should
	// convert as needed.
	//
	// JSON tag casing is inconsistent because the API itself is —
	// some fields use snake_case (client_id, domain_id), others
	// use lowercase-no-underscore (dateregistered, datebilleduntil).
	// Tags reflect what the API returns, not what's stylistically
	// uniform.
	Domain struct {
		ID                     string `json:"domain_id"`
		Domain                 string `json:"domain"`
		State                  string `json:"state"`
		API                    string `json:"api"`
		ClientID               string `json:"client_id"`
		ClientName             string `json:"client_name"`
		Locked                 string `json:"locked"`
		Private                string `json:"private"`
		Pending                string `json:"pending"`
		Premium                string `json:"premium"`
		RegistrantName         string `json:"registrant_name"`
		RegID                  string `json:"reg_id"`
		RegName                string `json:"reg_name"`
		AdmID                  string `json:"adm_id"`
		AdmName                string `json:"adm_name"`
		TecID                  string `json:"tec_id"`
		TecName                string `json:"tec_name"`
		AutoRenewTerm          string `json:"autorenew_term"`
		AutoRenewDaysRemaining string `json:"autorenew_days_remaining"`
		DateRegistered         string `json:"dateregistered"`
		DateModified           string `json:"datemodified"`
		DateBilledUntil        string `json:"datebilleduntil"`
		DateCancelled          string `json:"datecancelled"`
		DateLocked             string `json:"datelocked"`
	}

	// ListDomainsResponse represents the response from list_domains.
	// The API returns a paginated wrapper around the data array.
	ListDomainsResponse struct {
		Return struct {
			models.Pagination
			Data []Domain `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// DomainDetail is the full per-domain shape returned by
	// get_domain. Fields are PascalCase JSON keys matching the
	// underlying registry schema. RState is a numeric status
	// code; Locked / Private / Pending / Premium are real bools.
	DomainDetail struct {
		ClientID               int    `json:"ClientID"`
		Domain                 string `json:"Domain"`
		State                  string `json:"State"`
		RState                 int    `json:"RState"`
		AutorenewReminderSent  bool   `json:"autorenew_reminder_sent"`
		TransferAutorenewSent  bool   `json:"transfer_autorenew_sent"`
		API                    string `json:"API"`
		RegistrantName         string `json:"RegistrantName"`
		DateRegistered         string `json:"DateRegistered"`
		DateModified           string `json:"DateModified"`
		DateBilledUntil        string `json:"DateBilledUntil"`
		DateCancelled          string `json:"DateCancelled"`
		DatePrebilled          string `json:"dateprebilled"`
		DateLocked             string `json:"DateLocked"`
		DateRenewed            string `json:"daterenewed"`
		AutorenewTerm          int    `json:"autorenew_term"`
		AutorenewDaysRemaining int    `json:"autorenew_days_remaining"`
		RegistrantContactID    int    `json:"RegistrantContactID"`
		AdminContactID         int    `json:"AdminContactID"`
		TechnicalContactID     int    `json:"TechnicalContactID"`
		BillingContactID       int    `json:"BillingContactID"`
		Locked                 bool   `json:"Locked"`
		Private                bool   `json:"Private"`
		Pending                bool   `json:"Pending"`
		TransferStatus         string `json:"TransferStatus"`
		TransferID             int    `json:"TransferID"`
		AuthCodeGenerated      string `json:"auth_code_generated"`
		DateAdded              string `json:"date_added"`
		DateUpdated            string `json:"date_updated"`
		Premium                bool   `json:"premium"`
	}

	// GetDomainResponse represents the response from get_domain.
	GetDomainResponse struct {
		Return DomainDetail `json:"return"`
		models.APIResponse
	}

	// DomainPrice describes pricing for a single domain.
	DomainPrice struct {
		DomainPrice        float64 `json:"DomainPrice"`
		TotalPrice         float64 `json:"total_price"`
		TieredPrice        string  `json:"tiered_price"`
		BasePrice          string  `json:"base_price"`
		Premium            bool    `json:"premium"`
		BasePrivacyPrice   string  `json:"base_privacy_price"`
		TieredPrivacyPrice string  `json:"tiered_privacy_price"`
	}

	// GetDomainPriceResponse represents the response from
	// get_domain_price.
	GetDomainPriceResponse struct {
		Return DomainPrice `json:"return"`
		models.APIResponse
	}

	// CanTransferDomainInfo describes transfer eligibility.
	CanTransferDomainInfo struct {
		Domain      string `json:"domain"`
		CanTransfer bool   `json:"can_transfer"`
		Reason      string `json:"reason"`
	}

	// CanTransferDomainResponse represents the response from
	// can_transfer_domain.
	CanTransferDomainResponse struct {
		Return CanTransferDomainInfo `json:"return"`
		models.APIResponse
	}

	// BoolResponse is the response shape used by domain_available
	// and domain_inside_grace_period — Return is just a bool.
	BoolResponse struct {
		Return bool `json:"return"`
		models.APIResponse
	}

	// JobResponse is the standard response shape for SRS write
	// operations that queue a scheduler job — create_domain,
	// cancel_domain, etc. (Note: this matches the JobResponse
	// pattern used in other gosh packages, but is package-local
	// to keep srs self-contained.)
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
