package db

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates a new cloud database via /cloud/db/add.json.
//
// Required fields:
//   - ServerName: the CCS to create the database on.
//   - MySQLHost: the **DB stack name** on the CCS — e.g.
//     "mariadb1108", "mysql57", "postgres15". Despite the field
//     name, this is NOT a DNS hostname; it's the docker-stack
//     name resolvable inside the CCS's docker network.
//     Discover what's available via cloud.stack.List on the CCS
//     and match well-known prefixes (mariadb*/mysql*/postgres*).
//   - Database: the name of the new database (snake_case OK).
//   - Container: the name of an existing www-type container stack
//     on the same CCS that "owns" this database (used by the
//     Control Panel UI to associate the DB with its consumer).
//
// **Container-level lock between consecutive Adds** (verified live,
// May 2026): two cloud.db.Add calls against the **same Container**
// in quick succession reject the second with:
//
//	"Unable to create database. There is already a job operating
//	 on the container."
//
// even after the first call's scheduler job reports Completed. The
// container holds a brief follow-on lock past the Add's scheduler
// job. Insert ~30s of backoff between Adds against the same
// Container, OR pin each DB to a different Container, OR retry on
// the substring "already a job operating on the container."
//
// See examples/cloud-db-compare for the working pattern.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/db/add.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"database",
		"container",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("database", request.Database)
	values.Add("database", request.Database)
	values.Add("container", request.Container)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
