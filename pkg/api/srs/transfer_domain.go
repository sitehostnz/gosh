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
	values, err := encodeTransferDomain(opt)
	if err != nil {
		return response, err
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

func encodeTransferDomain(opt TransferDomainOptions) (url.Values, error) {
	values := url.Values{}
	values.Set("domain", opt.Domain)
	setNonEmpty(values, "udai", opt.UDAI)
	setNonZero(values, "params[registrant_contact_id]", opt.RegistrantContactID)
	setNonZero(values, "params[admin_contact_id]", opt.AdminContactID)
	setNonZero(values, "params[technical_contact_id]", opt.TechnicalContactID)
	setNonZero(values, "params[billing_contact_id]", opt.BillingContactID)
	setNonZero(values, "params[term]", opt.Term)
	if err := encodeNameServersParam(values, opt.NameServers); err != nil {
		return nil, err
	}
	return values, nil
}

func encodeNameServersParam(values url.Values, nss []NameServerEntry) error {
	for i, ns := range nss {
		if ns.Name == "" {
			return fmt.Errorf("srs.TransferDomain: NameServers[%d].Name is required", i)
		}
		idx := strconv.Itoa(i)
		values.Set("params[nameservers]["+idx+"][name]", ns.Name)
		setNonEmpty(values, "params[nameservers]["+idx+"][ipv4addr]", ns.IPv4Addr)
		setNonEmpty(values, "params[nameservers]["+idx+"][ipv6addr]", ns.IPv6Addr)
	}
	return nil
}

func setNonEmpty(values url.Values, key, val string) {
	if val != "" {
		values.Set(key, val)
	}
}

func setNonZero(values url.Values, key string, val int) {
	if val != 0 {
		values.Set(key, strconv.Itoa(val))
	}
}
