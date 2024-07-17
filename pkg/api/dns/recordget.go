package dns

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/models"
)

// GetRecord returns a record by id.
func (s *Client) GetRecord(ctx context.Context, request RecordRequest) (response models.DNSRecord, err error) {
	records, err := s.ListRecords(ctx, ListRecordsRequest{Domain: request.DomainName})
	if err != nil {
		return response, err
	}

	for _, record := range records.Return {
		if record.ID == request.ID {
			return record, nil
		}
	}

	return response, nil
}

// GetRecordWithType gets the record and filter based on type, as we often want to deal with the records of a specific type, mainly for use in external contexts, ie: terrorform.
func (s *Client) GetRecordWithType(ctx context.Context, request RecordRequest) (response []models.DNSRecord, err error) {
	records, err := s.ListRecords(ctx, ListRecordsRequest{Domain: request.DomainName})
	if err != nil {
		return response, err
	}

	var filteredRecords []models.DNSRecord
	for _, record := range records.Return {
		// Domain is not set in result from sitehost, we add it back in, but will ignore for now...
		if record.Type == request.RRType /*&& record.Domain == domain.Name*/ {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return filteredRecords, nil
}

// GetRecordWithRecord is a special case, for mainly when we are creating records where we need to get back what we just created.
//
// Deprecated: This is only needed for older versions of the Terraform provider as we now get a record ID back.
func (s *Client) GetRecordWithRecord(ctx context.Context, request models.DNSRecord) (response models.DNSRecord, err error) {
	records, err := s.ListRecords(ctx, ListRecordsRequest{Domain: request.Domain})
	if err != nil {
		return response, err
	}

	for _, r := range records.Return {
		if r.Name == request.Name &&
			r.Type == request.Type &&
			r.Content == request.Content &&
			(r.Priority == request.Priority || r.Priority == "0") {
			return r, nil
		}
	}
	return response, nil
}
