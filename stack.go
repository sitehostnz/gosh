package gosh

import (
	"context"
	"fmt"
)

type StackService service

type Container struct {
	Name        string `json:"name"`
	ContainerId string `json:"container_id"`
	State       string `json:"state"`
	Size        string `json:"size"`
	DateCreated string `json:"date_created"`

	SslEnabled bool `json:"ssl_enabled"`

	IsMissing interface{} `json:"is_missing"`
	Pending   interface{} `json:"pending"`
}

type Stack struct {
	ClientID    string `json:"client_id"`
	ServerID    string `json:"server_id"`
	Server      string `json:"server_name"`
	ServerLabel string `json:"server_label"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	DockerFile  string `json:"docker_file"`
	ServerOwner bool   `json:"server_owner"`
	IpAddress   string `json:"ip_addr_server"`
	DateAdded   string `json:"date_added"`
	DateUpdated string `json:"date_updated"`

	Containers *[]Container `json:"containers"`

	Pending   interface{} `json:"pending"`
	IsMissing interface{} `json:"is_missing"`
}

type StackResponse struct {
	Stack   *Stack `json:"return"`
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

type StackListResponse struct {
	Return struct {
		Stacks *[]Stack `json:"data"`
	}
	Status  bool   `json:"status"`
	Message string `json:"msg"`
}

func (s *StackService) Get(ctx context.Context, stack *Stack) (*Stack, error) {

	u := "cloud/stack/get.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"name",
	}

	v := req.URL.Query()
	v.Add("server", stack.Server)
	v.Add("name", stack.Name)

	req.URL.RawQuery = Encode(v, keys)

	response := new(StackResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Stack, nil
}

func (s *StackService) List(ctx context.Context, serverName string) (*[]Stack, error) {

	u := fmt.Sprintf("cloud/stack/list_all.json?filters[server_name]=%v", serverName)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(StackListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	//sort.Slice(*response.DomainRecords, func(i, j int) bool {
	//	return (*response.DomainRecords)[i].ChangeDate > (*response.DomainRecords)[j].ChangeDate
	//})
	//
	//// Add these back in for the sake of some sort of consistency
	//for i := range *response.DomainRecords {
	//	(*response.DomainRecords)[i].Domain = domain.Name
	//	(*response.DomainRecords)[i].ClientID = domain.ClientID
	//}
	//
	//return response.DomainRecords, err
	return response.Return.Stacks, nil
}
