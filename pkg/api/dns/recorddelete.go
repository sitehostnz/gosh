package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Delete a DNSRecord with the provided domain name.
func (s *Client) Delete(ctx context.Context, opts DeleteRecordRequest) (response models.APIResponse, err error) {
	u := "dns/delete_record.json"

	keys := []string{
		"client_id",
		"domain",
		"record_id",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", opts.Domain)
	values.Add("record_id", opts.RecordID)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
