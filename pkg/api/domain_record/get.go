package domain_record

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) Get(ctx context.Context, request GetRequest) (*models.DomainRecord, error) {
	records, err := s.List(ctx, ListRequest{DomainName: request.DomainName})
	if err != nil {
		return nil, err
	}

	for _, record := range *records {
		if record.Id == request.Id {
			return &record, nil
		}
	}
	return nil, nil
}

// get the record and filter based on type, as we often want to deal with the records of a specific type
// from external contexts, ie: terrorform
func (s *Client) GetWithType(ctx context.Context, request GetRequest) (*[]models.DomainRecord, error) {
	records, err := s.List(ctx, ListRequest{DomainName: request.DomainName})
	if err != nil {
		return nil, err
	}

	var filteredRecords []models.DomainRecord
	for _, record := range *records {
		// Domain is not set in result from sitehost, we add it back in, but will ignore for now...
		if record.Type == request.RRType /*&& record.Domain == domain.Name*/ {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return &filteredRecords, nil
}

// This is a special case, for mainly when we are creating records and we want to get back what we just created
func (s *Client) GetWithRecord(ctx context.Context, record models.DomainRecord) (*models.DomainRecord, error) {
	records, err := s.List(ctx, ListRequest{DomainName: record.Domain})
	if err != nil {
		return nil, err
	}

	for _, r := range *records {

		if r.Name == record.Name &&
			//TODO:  domain is not returned in the list
			// r.Domain == record.Domain &&
			// is only needed for creation

			// TODO: client id is not returned in list response
			// is only needed for creation
			// r.ClientID == record.ClientID &&
			// no idea what the state flag/field/property is for
			//r.State == record.State
			// ttl is not a per record here...
			// r.TTL == record.TTL &&

			r.Type == record.Type &&
			r.Content == record.Content &&
			r.Priority == record.Priority {
			return &r, nil
		}
	}
	return nil, nil

}
