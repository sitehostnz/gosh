package gosh

import (
	"context"
	"fmt"
)

type DatabaseUserService service

type DatabaseUser struct {
}

type DatabaseUserListResponse struct {
	Return struct {
		DatabaseUsers *[]DatabaseUser `json:"data"`
	}
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

func (s *DatabaseUserService) List(ctx context.Context) (*[]DatabaseUser, error) {
	u := fmt.Sprintf("cloud/db/user/list_all.json")
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(DatabaseUserListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Return.DatabaseUsers, err
}
