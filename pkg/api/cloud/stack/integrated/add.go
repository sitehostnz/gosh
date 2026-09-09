package integrated

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates an integrated (database) stack on a server, via
// "cloud/stack/integrated/add.json". It returns a job to poll.
//
// # Name is the image code, and the set is closed
//
// A database stack's name is its image code — "mariadb1108",
// "mysql57", and so on. Anything else is rejected with "invalid label
// passed. Please specify one from the following", which lists them.
// Those same names are the ones cloud/stack/add.json refuses as
// reserved: reserved *for* this endpoint, not reserved from you.
//
// The Stack* constants name the ones observed; discovering them from
// cloud/stack/image/list_all.json — the images whose
// nz.sitehost.image.type label is "integrated" — is more durable than
// any list compiled here.
//
// # Nothing else is configurable
//
// There is no label, SSL flag, compose file or environment to pass.
// The platform generates the compose itself from the image's latest
// version, and it has to: the file carries an env_file holding the
// database's root password and loopback port bindings derived from the
// image's own metadata. That is why building a database stack by hand
// through cloud/stack/add.json cannot work even with the port
// collisions resolved.
//
// # What it rejects
//
//   - a name outside the accepted set, as above;
//   - a server that is not on: "Unable to add stack, the server must
//     be on";
//   - a server that is locked;
//   - a stack of that name already present: "you already have a
//     mariadb1108 stack running" — so this is not idempotent, and
//     [Client.List] is the check to make first;
//   - a server belonging to another client.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	if request.ServerName == "" {
		return response, fmt.Errorf("cloud/stack/integrated.Add: ServerName is required")
	}
	if request.Name == "" {
		return response, fmt.Errorf("cloud/stack/integrated.Add: Name is required (the stack's image code, e.g. %q)", StackMariaDB1108)
	}

	keys := []string{"client_id", "server", "name"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("name", request.Name)

	req, err := s.client.NewRequest("POST", "cloud/stack/integrated/add.json", net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
