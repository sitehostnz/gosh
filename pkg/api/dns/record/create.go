package record

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Create a DNSRecord.
func (s *Client) Create(ctx context.Context, opts CreateRequest) (response models.APIResponse, err error) {
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
