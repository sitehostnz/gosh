package srs

import (
	"context"
)

// ListContacts the domain name contacts
func (s *Client) ListContacts(ctx context.Context) (response ListContactsResponse, err error) {
	u := "srs/list_contacts.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
