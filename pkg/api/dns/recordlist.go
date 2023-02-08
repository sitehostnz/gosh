package dns

// import (
// 	"context"

// 	"github.com/sitehostnz/gosh/pkg/models"
// )

// // Get returns record for a given domain.
// func (s *Client) Get(ctx context.Context, request RecordRequest) (*models.DNSRecord, error) {
// 	zone, err := s.GetZone(ctx, GetZoneRequest{DomainName: request.DomainName})
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, record := range *records {
// 		if record.ID == request.ID {
// 			return &record, nil
// 		}
// 	}

// 	return nil, nil
// }

// // GetWithType get the record and filter based on type, as we often want to deal with the records of a specific type, mainly for use in external contexts, ie: terrorform.
// func (s *Client) GetWithType(ctx context.Context, request RecordRequest) (*[]models.DNSRecord, error) {
// 	records, err := s.GetZone(ctx, GetZoneRequest{DomainName: request.DomainName})
// 	if err != nil {
// 		return nil, err
// 	}

// 	var filteredRecords []models.DNSRecord
// 	for _, record := range *records {
// 		// Domain is not set in result from sitehost, we add it back in, but will ignore for now...
// 		if record.Type == request.RRType /*&& record.Domain == domain.Name*/ {
// 			filteredRecords = append(filteredRecords, record)
// 		}
// 	}

// 	return &filteredRecords, nil
// }

// // GetWithRecord This is a special case, for mainly when we are creating records and we want to get back what we just created.
// func (s *Client) GetWithRecord(ctx context.Context, record models.DNSRecord) (*models.DNSRecord, error) {
// 	records, err := s.GetZone(ctx, GetZoneRequest{DomainName: record.Domain})
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, r := range *records {
// 		if r.Name == record.Name && r.Type == record.Type && r.Content == record.Content && r.Priority == record.Priority {
// 			return &r, nil
// 		}
// 	}

// 	return nil, nil
// }
