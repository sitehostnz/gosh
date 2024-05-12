package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// AddRecord adds a new record to the DNS for a domain.
// This function takes a context.Context and an AddRecordRequest struct as input parameters.
// It returns an APIResponse struct and an error.
func (s *Client) AddRecord(ctx context.Context, opts AddRecordRequest) (response AddRecordResponse, err error) {
	u := "dns/add_record.json"

	keys := []string{
		"client_id",
		"domain",
		"type",
		"name",
		"content",
		"prio",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", opts.Domain)
	values.Add("type", opts.Type)
	values.Add("name", opts.Name)
	values.Add("content", opts.Content)
	values.Add("prio", opts.Priority)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
