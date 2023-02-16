package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// UpdateRecord updates an existing DNS record for a domain.
// This function takes a context.Context and an UpdateRecordRequest struct as input parameters.
// It returns an APIResponse struct and an error.
func (s *Client) UpdateRecord(ctx context.Context, opts UpdateRecordRequest) (response models.APIResponse, err error) {
	u := "dns/update_record.json"

	keys := []string{
		"client_id",
		"domain",
		"record_id",
		"type",
		"name",
		"content",
		"priority",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", opts.Domain)
	values.Add("record_id", opts.RecordID)
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
