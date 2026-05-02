package srs

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListContacts returns all contacts for the authenticated client
// via "srs/list_contacts.json".
func (s *Client) ListContacts(ctx context.Context) (response ListContactsResponse, err error) {
	req, err := s.client.NewRequest("GET", "srs/list_contacts.json", "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// GetContact returns the full record for a single contact via
// "srs/get_contact.json". ContactID is required.
func (s *Client) GetContact(ctx context.Context, opt ContactOptions) (response GetContactResponse, err error) {
	if opt.ContactID == "" {
		return response, fmt.Errorf("srs.GetContact: ContactID is required")
	}
	path, err := net.AddOptions("srs/get_contact.json", opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// SearchContacts returns contacts matching the given filters
// via "srs/search_contacts.json". At least one of Name, Email,
// or RegistrantName must be set.
func (s *Client) SearchContacts(ctx context.Context, opt SearchContactsOptions) (response SearchContactsResponse, err error) {
	if opt.Name == "" && opt.Email == "" && opt.RegistrantName == "" {
		return response, fmt.Errorf("srs.SearchContacts: at least one of Name / Email / RegistrantName is required")
	}
	path, err := net.AddOptions("srs/search_contacts.json", opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
