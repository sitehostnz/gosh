package gosh

import (
	"context"
	"fmt"
	"net/url"
)

type DomainService service

type Domain struct {
	Name       string `json:"name"`
	ClientID   string `json:"client_id"`
	TemplateID string `json:"template_id"`
}

type DomainsListResponse struct {
	Return struct {
		Domains *[]Domain `json:"data"`
	}
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

type DomainRequest struct {
	ClientID string `url:"client_id" json:"client_id"`
	Domain   string `url:"domain" json:"domain"`
	// having an id here would be good, but the api does not expose them
}

type DomainResponse struct {
	Message string  `json:"msg"`
	Status  bool    `json:"status"`
	Time    float32 `json:"time"`
}

// List domains
// would be nice if we could explicitly ask for a single domain ...
// but the only way is filtering
func (s *DomainService) List(ctx context.Context) (*[]Domain, error) {
	u := fmt.Sprintf("dns/list_domains.json")
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(DomainsListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Return.Domains, err
}

func (s *DomainService) Get(ctx context.Context, domain *Domain) (*Domain, error) {
	u := fmt.Sprintf("dns/list_domains.json?filters[domain]=%v", domain.Name)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(DomainsListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	if len(*response.Return.Domains) == 0 {
		return nil, nil
	}

	return &(*response.Return.Domains)[0], nil
}

func (s *DomainService) Create(ctx context.Context, domain *Domain) (*Domain, error) {
	u := "dns/create_domain.json"

	keys := []string{
		"client_id",
		"domain",
	}

	values := url.Values{}
	values.Add("client_id", domain.ClientID)
	values.Add("domain", domain.Name)

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

	domain = &Domain{
		Name:       domain.Name,
		ClientID:   domain.ClientID,
		TemplateID: "0",
	}

	return domain, nil
}

func (s *DomainService) Delete(ctx context.Context, domain *Domain) (*DomainResponse, error) {

	u := "dns/delete_domain.json"

	keys := []string{
		"client_id",
		"domain",
	}

	values := url.Values{}
	values.Add("client_id", domain.ClientID)
	values.Add("domain", domain.Name)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(DomainResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return nil, nil
}