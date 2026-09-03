package db

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List returns cloud databases, via "cloud/db/list_all.json".
//
// # Filters are optional, but a bad one is not
//
// With no options this returns every database on the account, across
// every cloud server. It does not require a server name.
//
// A server name that does not resolve is a different matter: it is
// rejected with "Please specify a valid server name filter." rather
// than returning an empty page. So an empty result means the server
// genuinely has no databases, and never that the filter was ignored —
// which is what makes filtering here safe to rely on.
//
// MySQLHost filters by the database host container ("mariadb0"), and
// Database by exact name. Results are paged; see [models.Filtering].
func (s *Client) List(ctx context.Context, opt ListOptions) (response ListResponse, err error) {
	uri := "cloud/db/list_all.json"

	path, err := net.AddOptions(uri, opt)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
