package srs

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
	"github.com/sitehostnz/gosh/pkg/models"
)

// CreateContactOptions describes a new domain contact for the
// Shared Registry System (POST srs/create_contact.json).
//
// # Wire schema (per docs.sitehost.nz, May 2026)
//
// The endpoint is form-encoded. Two naming conventions coexist:
//
//   - Top-level (no params[] wrapper): name, email, postal_address,
//     postal_address2, suburb, city, country.
//   - params[<PascalCase>] for everything else: PostCode, Province,
//     Organisation (see field name on the struct for the wire key).
//   - params[<Group>][<Field>] for the three phone-style numbers:
//     params[Phone][Country|Area|Local|Extension] and identical
//     shapes for Fax and Mobile.
//
// Field-name capitalisation matters. Sub-keys are PascalCase
// (Country, not country / cntry / CountryCode). Lowercase or
// abbreviated variants are silently rejected with confusing
// errors like "phone country code number is missing" or "phone
// number is missing".
//
// # Required vs optional (live finding — diverges from public docs)
//
// Required (server-side validated):
//   - Name, Email
//   - PostalAddress, PostalAddress2
//   - Suburb (live: omitting it returns "The suburb is missing.")
//   - City, Country (ISO 2-letter, e.g. "NZ")
//   - PostCode (params[PostCode])
//   - Phone: Country, Area, Local (Extension optional)
//
// Optional but commonly required by registries:
//   - Province, Organisation
//   - Fax, Mobile (full sub-arrays)
//
// # Email TLD constraint (live finding)
//
// Email must use a real public-suffix TLD (.com, .nz, …).
// RFC-2606 reserved TLDs (.example, .test, .invalid) are rejected
// with "Please specify a valid email address." Use example.com,
// not example.example.
//
// # Phone format
//
// Country: digits only, e.g. "64" for NZ.
// Area: digits only, no leading zero ("9" not "09" — though the
// API accepts and normalises "09").
// Local: line number digits only, no separators.
// Extension: optional digits.
type CreateContactOptions struct {
	Name           string `url:"name"`
	Email          string `url:"email"`
	PostalAddress  string `url:"postal_address"`
	PostalAddress2 string `url:"postal_address2"`
	Suburb         string `url:"suburb"`
	City           string `url:"city"`
	Country        string `url:"country"`
	PostCode       string `url:"params[PostCode]"`
	Province       string `url:"params[Province],omitempty"`
	Organization   string `url:"params[Organization],omitempty"` //nolint:misspell // matches the upstream API wire field name

	// Phone is required. Per the published schema: Country, Area,
	// Local are required; Extension is optional.
	PhoneCountry   string `url:"params[Phone][Country],omitempty"`
	PhoneArea      string `url:"params[Phone][Area],omitempty"`
	PhoneLocal     string `url:"params[Phone][Local],omitempty"`
	PhoneExtension string `url:"params[Phone][Extension],omitempty"`

	// Fax — fully optional. Same Country/Area/Local/Extension shape.
	FaxCountry   string `url:"params[Fax][Country],omitempty"`
	FaxArea      string `url:"params[Fax][Area],omitempty"`
	FaxLocal     string `url:"params[Fax][Local],omitempty"`
	FaxExtension string `url:"params[Fax][Extension],omitempty"`

	// Mobile — fully optional. Same Country/Area/Local/Extension
	// shape. Area on NZ mobiles is the network prefix (e.g. "21").
	MobileCountry   string `url:"params[Mobile][Country],omitempty"`
	MobileArea      string `url:"params[Mobile][Area],omitempty"`
	MobileLocal     string `url:"params[Mobile][Local],omitempty"`
	MobileExtension string `url:"params[Mobile][Extension],omitempty"`
}

// CreateContactResponse returns the new contact's id (and a
// snapshot of the contact body the registry stored).
//
// The JSON key is "ContactID" (PascalCase) per the live API.
type CreateContactResponse struct {
	Return struct {
		ContactID string `json:"ContactID"`
	} `json:"return"`
	models.APIResponse
}

// CreateContact registers a new domain contact via
// "srs/create_contact.json". Returns the new contact's id.
//
// Contact fields outside the required set may be left empty; a
// follow-up UpdateContact can fill them in.
func (s *Client) CreateContact(ctx context.Context, opt CreateContactOptions) (response CreateContactResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("srs.CreateContact: Name is required")
	}
	if opt.Email == "" {
		return response, fmt.Errorf("srs.CreateContact: Email is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/create_contact.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// UpdateContactOptions edits an existing contact. Only the fields
// you set are sent (omitempty everywhere). All field names are
// PascalCase per the API's params[<Name>] convention.
type UpdateContactOptions struct {
	ContactID      int    `url:"contact_id"`
	Name           string `url:"params[Name],omitempty"`
	Email          string `url:"params[Email],omitempty"`
	PostalAddress  string `url:"params[PostalAddress],omitempty"`
	PostalAddress2 string `url:"params[PostalAddress2],omitempty"`
	Suburb         string `url:"params[Suburb],omitempty"`
	City           string `url:"params[City],omitempty"`
	Country        string `url:"params[Country],omitempty"`
	PostCode       string `url:"params[PostCode],omitempty"`
	Province       string `url:"params[Province],omitempty"`
}

// UpdateContact edits an existing domain contact via
// "srs/update_contact.json". ContactID is required; only the
// non-zero / non-empty params[*] fields are sent.
func (s *Client) UpdateContact(ctx context.Context, opt UpdateContactOptions) (response models.APIResponse, err error) {
	if opt.ContactID == 0 {
		return response, fmt.Errorf("srs.UpdateContact: ContactID is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/update_contact.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// DeleteContactOptions identifies a contact to remove.
type DeleteContactOptions struct {
	ContactID int `url:"contact_id"`
}

// DeleteContact removes a domain contact via
// "srs/delete_contact.json". The API rejects deletion of contacts
// currently bound to any domain — unbind first via
// UpdateDomainContacts.
func (s *Client) DeleteContact(ctx context.Context, opt DeleteContactOptions) (response models.APIResponse, err error) {
	if opt.ContactID == 0 {
		return response, fmt.Errorf("srs.DeleteContact: ContactID is required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/delete_contact.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// UpdateDomainContactsOptions rebinds the four contact roles
// (registrant / admin / technical / billing) on a domain. All
// four contact IDs are required by the API except BillingContact
// per the docs (which marks billing as optional, but in practice
// most TLDs require it).
type UpdateDomainContactsOptions struct {
	Domain              string `url:"domain"`
	RegistrantContactID int    `url:"registrant_contact_id"`
	AdminContactID      int    `url:"admin_contact_id"`
	TechnicalContactID  int    `url:"technical_contact_id"`
	BillingContactID    int    `url:"billing_contact_id,omitempty"`
}

// UpdateDomainContacts rebinds the contact roles on a domain via
// "srs/update_domain_contacts.json".
//
// Note .nz registry policy may restrict registrant-name changes
// (the registrant is the legal owner; transferring it elsewhere
// is a special operation, not just a contact update). Other
// roles (admin, tech, billing) typically rebind freely.
func (s *Client) UpdateDomainContacts(ctx context.Context, opt UpdateDomainContactsOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.UpdateDomainContacts: Domain is required")
	}
	if opt.RegistrantContactID == 0 || opt.AdminContactID == 0 || opt.TechnicalContactID == 0 {
		return response, fmt.Errorf("srs.UpdateDomainContacts: RegistrantContactID, AdminContactID, TechnicalContactID are all required")
	}
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/update_domain_contacts.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// UpdateCompanyInfoOptions updates account-level company info
// (used in registry whois output and renewal email branding).
// All fields optional — only non-empty fields are sent. Use
// GetCompanyInfo to read current values before writing.
type UpdateCompanyInfoOptions struct {
	CompanyName          string `url:"params[CompanyName],omitempty"`
	CompanyURL           string `url:"params[CompanyUrl],omitempty"`
	CompanyRenewURL      string `url:"params[CompanyRenewUrl],omitempty"`
	CompanyEmail         string `url:"params[CompanyEmail],omitempty"`
	CompanyEmailFrom     string `url:"params[CompanyEmailFrom],omitempty"`
	CompanyEmailFromName string `url:"params[CompanyEmailFromName],omitempty"`
	CompanySupportEmail  string `url:"params[CompanySupportEmail],omitempty"`
	CompanyPhone         string `url:"params[CompanyPhone],omitempty"`
	CompanyFax           string `url:"params[CompanyFax],omitempty"`
	// SendRenewedEmail is "1" for yes, "0" for no. Stringified to
	// distinguish "unset" (don't send) from explicit-zero ("0",
	// disable).
	SendRenewedEmail string `url:"params[SendRenewedEmail],omitempty"`
}

// UpdateCompanyInfo updates account-level company info via
// "srs/update_company_info.json".
//
// The endpoint affects branding shown in registry whois output
// and the renewal-email templates. Test against this carefully —
// mistyped values affect how customers see the account.
//
// Best practice: read current values via srs.GetCompanyInfo
// first, copy them into UpdateCompanyInfoOptions, change only
// what you need to change, send the rest unchanged.
func (s *Client) UpdateCompanyInfo(ctx context.Context, opt UpdateCompanyInfoOptions) (response models.APIResponse, err error) {
	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "srs/update_company_info.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
