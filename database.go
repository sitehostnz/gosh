package gosh

import (
	"context"
	"fmt"
)

type DatabaseService service

type Database struct {
	Id          string      `json:"id"`
	DbName      string      `json:"db_name"`
	MysqlHost   string      `json:"mysql_host"`
	Size        string      `json:"size"`
	ClientId    string      `json:"client_id"`
	ServerId    string      `json:"server_id"`
	Pending     interface{} `json:"pending"`
	IsMissing   string      `json:"is_missing"`
	DateAdded   string      `json:"date_added"`
	DateUpdated string      `json:"date_updated"`
	ServerName  string      `json:"server_name"`
	ServerLabel string      `json:"server_label"`
	ServerIp    string      `json:"server_ip"`
	ServerOwner bool        `json:"server_owner"`
	Container   string      `json:"container"`
}

type DatabaseListResponse struct {
	Return struct {
		Databases *[]Database `json:"data"`
	}
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

// List domains
// would be nice if we could explicitly ask for a single domain ...
// but the only way is filtering
func (s *DatabaseService) List(ctx context.Context) (*[]Database, error) {
	u := fmt.Sprintf("cloud/db/list_all.json")
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(DatabaseListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Return.Databases, err
}
