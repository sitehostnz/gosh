package gosh

import (
	"context"

	"github.com/go-resty/resty/v2"
)

const (
	serversName = "servers"

	serversEndpoint = "server"
)

type Resource struct {
	name     string
	endpoint string
	R        func(ctx context.Context) *resty.Request
	PR       func(ctx context.Context) *resty.Request
}

func NewResource(client *Client, name string, endpoint string, singleType interface{}, pagedType interface{}) *Resource {
	r := func(ctx context.Context) *resty.Request {
		return client.R(ctx).SetResult(singleType)
	}

	pr := func(ctx context.Context) *resty.Request {
		return client.R(ctx).SetResult(pagedType)
	}

	return &Resource{name, endpoint, r, pr}
}

func (r Resource) Endpoint() (string, error) {
	return r.endpoint, nil
}
