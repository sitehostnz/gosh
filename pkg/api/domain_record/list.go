package domain_record

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/models"
	"sort"
)

func (s *Client) List(ctx context.Context, request ListRequest) (*[]models.DomainRecord, error) {

	u := fmt.Sprintf("dns/list_records.json?client_id=%v&domain=%v", s.client.ClientID, request.DomainName)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
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
