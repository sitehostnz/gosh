package domain_record

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

func (s *Client) Create(ctx context.Context, domainRecord *models.DomainRecord) (*models.DomainRecord, error) {

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
	values.Add("domain", domainRecord.Domain)
	values.Add("type", domainRecord.Type)
	values.Add("name", domainRecord.Name)
	values.Add("content", adjustContent(domainRecord))
	values.Add("prio", domainRecord.Priority)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(models.APIResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// it would be really nice if the create record could return us the ID of the newly created domain
	// will need to ask some people, the api will blindly create new records, even if they are identical in terms of
	// content, may be a fun time once we get to terraform land
	// as we are going to need to add the record, then get them. and assume the most recent one we get back
	// is the one we've just added ... so we're gunna go back and get it...
	record, err := s.GetWithRecord(ctx, *domainRecord)
	if err != nil {
		return nil, err
	}

	return record, nil

}
