package gosh

import (
	"context"
)

type DomainRecordService service

type DomainRecord struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	TTL        string `json:"ttl"`
	Priority   string `json:"prio"`
	ChangeDate string `json:"change_date"`
	State      string `json:"state"`
}

type DomainRecordResponse struct {
	Return struct {
		DomainRecords *[]DomainRecord `json:"return"`
	}
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

func (s *DomainRecordService) ListRecords(ctx context.Context, domain string) (*[]DomainRecord, error) {

	u := "dns/list_records.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(DomainRecordResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Return.DomainRecords, err
}

//CreateDomainRecord(context.Context, *DomainRecord) (*DomainRecord, *Response, error)
//DeleteDomainRecord(context.Context, string, string) (*Response, error)
