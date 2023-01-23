package record

import (
	"context"
	"fmt"
	"sort"

	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) GetZone(ctx context.Context, request ZoneRequest) (*[]models.DNSRecord, error) {

	u := fmt.Sprintf("dns/list_records.json?client_id=%v&domain=%v", s.client.ClientID, request.DomainName)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ZoneResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	sort.Slice(*response.DomainRecords, func(i, j int) bool {
		return (*response.DomainRecords)[i].ChangeDate > (*response.DomainRecords)[j].ChangeDate
	})

	// Add these back in for the sake of some sort of consistency
	for i := range *response.DomainRecords {
		(*response.DomainRecords)[i].Domain = request.DomainName
		(*response.DomainRecords)[i].ClientID = s.client.ClientID
	}

	return response.DomainRecords, err
}

func (s *Client) Get(ctx context.Context, request RecordRequest) (*models.DNSRecord, error) {
	records, err := s.GetZone(ctx, ZoneRequest{DomainName: request.DomainName})
	if err != nil {
		return nil, err
	}

	for _, record := range *records {
		if record.ID == request.Id {
			return &record, nil
		}
	}
	return nil, nil
}

// GetWithType get the record and filter based on type, as we often want to deal with the records of a specific type, mainly for use in external contexts, ie: terrorform.
func (s *Client) GetWithType(ctx context.Context, request RecordRequest) (*[]models.DNSRecord, error) {
	records, err := s.GetZone(ctx, ZoneRequest{DomainName: request.DomainName})
	if err != nil {
		return nil, err
	}

	var filteredRecords []models.DNSRecord
	for _, record := range *records {
		// Domain is not set in result from sitehost, we add it back in, but will ignore for now...
		if record.Type == request.RRType /*&& record.Domain == domain.Name*/ {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return &filteredRecords, nil
}

// GetWithRecord This is a special case, for mainly when we are creating records and we want to get back what we just created.
func (s *Client) GetWithRecord(ctx context.Context, record models.DNSRecord) (*models.DNSRecord, error) {
	records, err := s.GetZone(ctx, ZoneRequest{DomainName: record.Domain})
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
			// r.State == record.State
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
