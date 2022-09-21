package gosh

import (
	"context"
	"fmt"
)

type StackImageService service

type StackImage struct {
	Id             string      `json:"id"`
	ClientId       string      `json:"client_id"`
	Label          string      `json:"label"`
	Code           string      `json:"code"`
	DateAdded      string      `json:"date_added"`
	DateUpdated    string      `json:"date_updated"`
	IsPublic       bool        `json:"is_public"`
	IsMissing      bool        `json:"is_missing"`
	ProjectId      string      `json:"project_id"`
	RegistryId     string      `json:"registry_id"`
	ForkedFrom     interface{} `json:"forked_from"`
	Pending        interface{} `json:"pending"`
	IsCustom       bool        `json:"is_custom"`
	BuildStatus    string      `json:"build_status"`
	ClientName     string      `json:"client_name"`
	VersionCount   int         `json:"version_count"`
	ContainerCount int         `json:"container_count"`
}

type StackImageListResponse struct {
	Return struct {
		Images *[]StackImage `json:"data"`
	}
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

func (s *StackImageService) List(ctx context.Context) (*[]StackImage, error) {

	u := fmt.Sprintf("cloud/image/list_all.json")

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(StackImageListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Return.Images, nil
}
