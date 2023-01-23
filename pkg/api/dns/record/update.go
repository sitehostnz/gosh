package record

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

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
	// not sure as to the best way to do this, we are we likely to move records between clients?
	// lets just use the one specified in the config for now...
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", domainRecord.Domain)
	values.Add("record_id", domainRecord.ID)
	values.Add("type", domainRecord.Type)
	values.Add("name", domainRecord.Name)
	values.Add("content", adjustContent(domainRecord))
	values.Add("prio", domainRecord.Priority)

	// it would be really nice if the create record could return us the ID of the newly created domain
	// will need to ask some people, the api will blindly create new records, even if they are identical in terms of
	// content, may be a fun time once we get to terraform land
	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(models.APIResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// read back the updated record...
	domainRecord, err = s.GetWithRecord(ctx, *domainRecord)
	if err != nil {
		return nil, err
	}

	return domainRecord, nil

}
