package domain_record

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

func (s *Client) Delete(ctx context.Context, domainRecord *models.DomainRecord) (*models.DomainRecord, error) {
	u := "dns/delete_record.json"

	keys := []string{
		"client_id",
		"domain",
		"record_id",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", domainRecord.Domain)
	values.Add("record_id", domainRecord.ID)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	err = s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil

}
