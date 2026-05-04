package srs

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
)

// TransferDomainOptions submits a domain-transfer-in request via
// "srs/transfer_domain.json".
//
// **Cost warning:** transfers typically charge a renewal-priced
// term at the registry and add a year (or `params[term]` months)
// to the expiry. Confirm pricing via srs.GetDomainPrice with
// type="transfer" before calling.
//
// Required:
//   - Domain — the domain to transfer in.
//
// Conditionally required:
//   - UDAI — the registry auth code from the losing registrar.
//     Required for TLDs that use UDAI/auth-code transfers (.nz,
//     gTLDs). Some legacy registries don't require it.
//
// Optional but commonly required by registries:
//   - RegistrantContactID, AdminContactID, TechnicalContactID,
//     BillingContactID — defaults to the account's primary
//     contact when omitted, but most registries require an
//     explicit registrant on transfer-in.
//   - Term — months. Defaults to the registry minimum (typically
//     12).
//   - NameServers — initial nameserver set. Defaults to the
//     existing set carried over from the losing registrar.
type TransferDomainOptions struct {
	Domain              string
	UDAI                string
	RegistrantContactID int
	AdminContactID      int
	TechnicalContactID  int
	BillingContactID    int
	Term                int
	NameServers         []NameServerEntry
}

// TransferDomain initiates an inbound domain transfer via
// "srs/transfer_domain.json".
//
// Live-validating this endpoint requires a domain held at another
// registrar plus a valid UDAI — gosh's test suite stops at unit
// tests for this reason. The wrapper is intentionally minimal:
// it forwards the documented parameters as-is.
func (s *Client) TransferDomain(ctx context.Context, opt TransferDomainOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.TransferDomain: Domain is required")
	}
	values := url.Values{}
	values.Set("domain", opt.Domain)
	if opt.UDAI != "" {
		values.Set("udai", opt.UDAI)
	}
	if opt.RegistrantContactID != 0 {
		values.Set("params[registrant_contact_id]", strconv.Itoa(opt.RegistrantContactID))
	}
	if opt.AdminContactID != 0 {
		values.Set("params[admin_contact_id]", strconv.Itoa(opt.AdminContactID))
	}
	if opt.TechnicalContactID != 0 {
		values.Set("params[technical_contact_id]", strconv.Itoa(opt.TechnicalContactID))
	}
	if opt.BillingContactID != 0 {
		values.Set("params[billing_contact_id]", strconv.Itoa(opt.BillingContactID))
	}
	if opt.Term > 0 {
		values.Set("params[term]", strconv.Itoa(opt.Term))
	}
	for i, ns := range opt.NameServers {
		if ns.Name == "" {
			return response, fmt.Errorf("srs.TransferDomain: NameServers[%d].Name is required", i)
		}
		idx := strconv.Itoa(i)
		values.Set("params[nameservers]["+idx+"][name]", ns.Name)
		if ns.IPv4Addr != "" {
			values.Set("params[nameservers]["+idx+"][ipv4addr]", ns.IPv4Addr)
		}
		if ns.IPv6Addr != "" {
			values.Set("params[nameservers]["+idx+"][ipv6addr]", ns.IPv6Addr)
		}
	}
	req, err := s.client.NewRequest("POST", "srs/transfer_domain.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
