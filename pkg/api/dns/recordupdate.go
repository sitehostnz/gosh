package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Update a record for a given domain.
func (s *Client) Update(ctx context.Context, domainRecord *models.DNSRecord) (*models.DNSRecord, error) {
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
	values.Add("domain", domainRecord.Domain)
	values.Add("record_id", domainRecord.ID)
	values.Add("type", domainRecord.Type)
	values.Add("name", domainRecord.Name)
	values.Add("prio", domainRecord.Priority)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(models.APIResponse)
	if err := s.client.Do(ctx, req, &response); err != nil {
		return nil, err
	}

	return domainRecord, nil
}
