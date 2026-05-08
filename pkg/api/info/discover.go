package info

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/api"
)

// NewClientWithDiscovery returns a new *api.Client given just an
// API key, by calling api/get_info.json to discover the
// authenticated client's client_id.
//
// Use this when the caller does not yet know the client_id — for
// example a tool bootstrapping against a freshly-issued key. Use
// api.New(apiKey, clientID, opts...) directly if the client_id is
// known, including when targeting a sub-account from a super-user
// key (discovery resolves to the super-user's own id, which is
// not the sub-account's id).
//
// opts are forwarded to the returned client unchanged; if
// api.SetBaseURL is supplied it also applies to the discovery
// call. The placeholder "0" used internally for the bootstrap
// satisfies NewRequest's empty-check; api/get_info.json
// authenticates against the apikey alone and does not validate
// the client_id query parameter.
func NewClientWithDiscovery(ctx context.Context, apiKey string, opts ...api.ClientOpt) (*api.Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("info.NewClientWithDiscovery: apiKey must not be empty")
	}

	bootstrap, err := api.New(apiKey, "0", opts...)
	if err != nil {
		return nil, fmt.Errorf("info.NewClientWithDiscovery: bootstrap client: %w", err)
	}

	resp, err := New(bootstrap).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("info.NewClientWithDiscovery: api/get_info.json: %w", err)
	}
	if resp.Return.ClientID == "" {
		return nil, fmt.Errorf("info.NewClientWithDiscovery: api/get_info.json returned empty client_id")
	}

	return api.New(apiKey, resp.Return.ClientID, opts...)
}
