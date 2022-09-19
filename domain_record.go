package gosh

import (
	"context"
	"fmt"
	"net/url"
	"sort"
)

type DomainRecordService service

type DomainRecord struct {
	Id         string `json:"id"`
	ClientID   string `json:client_id`
	Name       string `json:"name"`
	Domain     string `json:domain`
	Type       string `json:"type"`
	Content    string `json:"content"`
	TTL        string `json:"ttl"`
	Priority   string `json:"prio"`
	ChangeDate string `json:"change_date"`
	State      string `json:"state"`
}

type DomainRecordResponse struct {
	DomainRecords *[]DomainRecord `json:"return"`
	Status        bool            `json:"status"`
	Message       string          `json:"msg"`
}

func (s *DomainRecordService) List(ctx context.Context, domain *Domain) (*[]DomainRecord, error) {

	u := fmt.Sprintf("dns/list_records.json?client_id=%v&domain=%v", domain.ClientID, domain.Name)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(DomainRecordResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	sort.Slice(*response.DomainRecords, func(i, j int) bool {
		return (*response.DomainRecords)[i].ChangeDate > (*response.DomainRecords)[j].ChangeDate
	})

	// Add these back in for the sake of some sort of consistency
	for i := range *response.DomainRecords {
		(*response.DomainRecords)[i].Domain = domain.Name
		(*response.DomainRecords)[i].ClientID = domain.ClientID
	}

	return response.DomainRecords, err
}

func (s *DomainRecordService) Get(ctx context.Context, domain *Domain, id string) (*DomainRecord, error) {
	records, err := s.List(ctx, domain)
	if err != nil {
		return nil, err
	}

	for _, record := range *records {
		if record.Id == id {
			return &record, nil
		}
	}
	return nil, nil
}

func (s *DomainRecordService) GetRecordsWithType(ctx context.Context, domain *Domain, rrtype string) (*[]DomainRecord, error) {
	records, err := s.List(ctx, domain)
	if err != nil {
		return nil, err
	}

	var filteredRecords []DomainRecord
	for _, record := range *records {
		// Domain is not set in result from sitehost, we add it back in, but will ignore for now...
		if record.Type == rrtype /*&& record.Domain == domain.Name*/ {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return &filteredRecords, nil
}

func (s *DomainRecordService) GetRecordWithRecord(ctx context.Context, record *DomainRecord) (*DomainRecord, error) {
	records, err := s.List(ctx, &Domain{
		Name:       record.Domain,
		ClientID:   record.ClientID,
		TemplateID: "0",
	})

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

			r.Type == record.Type &&
			r.Content == record.Content &&
			r.TTL == record.TTL &&
			r.Priority == record.Priority &&
			// no idea what the state flag/field/property is for
			r.State == record.State {
			return &r, nil
		}
	}
	return nil, nil
}

func (s *DomainRecordService) Add(ctx context.Context, domainRecord *DomainRecord) (*DomainRecord, error) {

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
	values.Add("client_id", domainRecord.ClientID)
	values.Add("domain", domainRecord.Domain)
	values.Add("type", domainRecord.Type)
	values.Add("name", domainRecord.Name)
	values.Add("content", domainRecord.Content)
	values.Add("prio", domainRecord.Priority)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(DomainResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	if !response.Status {
		return nil, fmt.Errorf("%s", response.Message)
	}

	// it would be really nice if the create record could return us the ID of the newly created domain
	// will need to ask some people, the api will blindly create new records, even if they are identical in terms of
	// content, may be a fun time once we get to terraform land
	// as we are going to need to add the record, then get them. and assume the most recent one we get back
	// is the one we've just added ... so we're gunna go back and get it...
	record, err := s.GetRecordWithRecord(ctx, domainRecord)
	if err != nil {
		return nil, err
	}

	return record, nil

}

func (s *DomainRecordService) Update(ctx context.Context, domainRecord *DomainRecord) (*DomainRecord, error) {

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
	values.Add("client_id", domainRecord.ClientID)
	values.Add("domain", domainRecord.Domain)
	values.Add("record_id", domainRecord.Id)
	values.Add("type", domainRecord.Type)
	values.Add("name", domainRecord.Name)
	values.Add("content", domainRecord.Content)
	values.Add("prio", domainRecord.Priority)

	// it would be really nice if the create record could return us the ID of the newly created domain
	// will need to ask some people, the api will blindly create new records, even if they are identical in terms of
	// content, may be a fun time once we get to terraform land
	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(DomainResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	if !response.Status {
		return nil, fmt.Errorf("%s", response.Message)
	}

	return domainRecord, nil

}

func (s *DomainRecordService) Delete(ctx context.Context, domainRecord *DomainRecord) (*DomainRecord, error) {
	u := "dns/delete_record.json"

	keys := []string{
		"client_id",
		"domain",
		"record_id",
	}

	values := url.Values{}
	values.Add("client_id", domainRecord.ClientID)
	values.Add("domain", domainRecord.Domain)
	values.Add("record_id", domainRecord.Id)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	err = s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil

}
