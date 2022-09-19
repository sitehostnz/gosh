package gosh

import (
	"context"
)

type ApiInfoService service

type ApiInfo struct {
	ClientID  string    `json:"client_id"`
	Modules   *[]string `json:"modules"`
	ContactID string    `json:"contact_id"`
	Roles     *[]string `json:"roles"`
}

type ApiInfoResponse struct {
	ApiInfo *ApiInfo `json:"return"`
	Status  bool     `json:"status"`
	Message string   `json:"msg"`
}

func (s *ApiInfoService) Get(ctx context.Context) (*ApiInfo, error) {

	req, err := s.client.NewRequest("GET", "api/get_info.json", "")
	if err != nil {
		return nil, err
	}

	response := new(ApiInfoResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.ApiInfo, nil
}
