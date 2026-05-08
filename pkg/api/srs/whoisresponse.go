package srs

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// WhoisContact is a contact block within a whois response. The
	// PostalAddress field is itself a key/value map because the
	// underlying registry returns a variable set of address fields
	// per domain (e.g. {"Country": "NZ"} for some, with Street /
	// City / Postcode for others) — using a map captures whatever
	// the API returns without forcing speculation about field names.
	WhoisContact struct {
		Name          string            `json:"Name"`
		Company       string            `json:"Company"`
		Email         string            `json:"Email"`
		PostalAddress map[string]string `json:"PostalAddress"`
	}

	// WhoisNameServer is a name-server entry within a whois response.
	WhoisNameServer struct {
		FQDN    string `json:"FQDN"`
		IP4Addr string `json:"IP4Addr"`
		IP6Addr string `json:"IP6Addr"`
	}

	// Whois is the structured whois result for a domain.
	// Field-name capitalisation in the API is unusual (PascalCase JSON
	// keys), reflecting the underlying registry response shape.
	Whois struct {
		Domain          string            `json:"DomainName"`
		State           string            `json:"Status"`
		DateRegistered  string            `json:"RegisteredDate"`
		DateModified    string            `json:"ModifiedDate"`
		DateBilledUntil string            `json:"BilledUntil"`
		NameServers     []WhoisNameServer `json:"NameServers"`
		Registrant      WhoisContact      `json:"RegistrantContact"`
		Technical       WhoisContact      `json:"TechnicalContact"`
		Admin           WhoisContact      `json:"AdminContact"`
	}

	// WhoisResponse represents the response from the whois endpoint.
	// SourceIP is the address the whois query was sourced from
	// (returned by the API alongside the registry data).
	WhoisResponse struct {
		Return   Whois  `json:"return"`
		SourceIP string `json:"SourceIP"`
		models.APIResponse
	}
)
