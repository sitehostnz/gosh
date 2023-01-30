package domain_record

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/models"
	"log"
	"sort"
)

func (s *Client) GetZone(ctx context.Context, request ZoneRequest) (*[]models.DomainRecord, error) {

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
	for i, record := range *response.DomainRecords {
		record.Domain = request.DomainName
		record.ClientID = s.client.ClientID
		record.Content = NormaliseContent(record)
		record.Name = DeconstructFqdn(record.Name, request.DomainName)
		(*response.DomainRecords)[i] = record

	}

	return response.DomainRecords, err
}

func (s *Client) Get(ctx context.Context, request RecordRequest) (*models.DomainRecord, error) {
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
func (s *Client) GetWithType(ctx context.Context, request RecordRequest) (*[]models.DomainRecord, error) {
	records, err := s.GetZone(ctx, ZoneRequest{DomainName: request.DomainName})
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

// GetWithRecord This is a special case, for mainly when we are creating records and we want to get back what we just created.
func (s *Client) GetWithRecord(ctx context.Context, record models.DomainRecord) (*models.DomainRecord, error) {
	records, err := s.GetZone(ctx, ZoneRequest{DomainName: record.Domain})
	if err != nil {
		return nil, err
	}

	for _, r := range *records {
		log.Printf("[INFO] Domain Record: %s %s", ConstructFqdn(r.Name, record.Domain), ConstructFqdn(record.Name, record.Domain))
		log.Printf("[INFO] Domain Record: %s %s", ConstructFqdn(r.Name, record.Domain), ConstructFqdn(record.Name, record.Domain))

		if ConstructFqdn(r.Name, record.Domain) == ConstructFqdn(record.Name, record.Domain) &&
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
			NormaliseContent(r) == NormaliseContent(record) &&
			r.Priority == record.Priority {
			return &r, nil
		}
	}
	return nil, nil

}
