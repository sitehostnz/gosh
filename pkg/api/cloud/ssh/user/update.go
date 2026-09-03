package user

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
	"github.com/sitehostnz/gosh/pkg/shtypes"
)

// Update changes an existing SSH user, via
// "cloud/ssh/user/update.json".
//
// # ReadOnlyConfig was silently ignored
//
// The value was added to the form as params[read_only_config] while
// the keys list named params[read_only_config][], with an array
// suffix. net.Encode emits only the keys it is given, so the value was
// dropped: setting ReadOnlyConfig did nothing, the call still
// succeeded, and nothing indicated it had been ignored.
//
// The scalar spelling is the one kept here, on the grounds that it is
// what [UpdateRequest] and the sibling [Client.Add] both use and that
// the value is a flag rather than a list. That reasoning is inference,
// not observation — the fix has not been confirmed against a live
// server, because doing so means changing a real user's access. Worth
// verifying the first time anyone uses it in anger.
//
// The three genuine list parameters around it — containers, ssh_keys,
// volumes — do carry the suffix, which is presumably where it came
// from.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/ssh/user/update.json"
	keys := []string{
		"client_id",
		"server_name",
		"username",
		"params[password]",
		"params[containers][]",
		"params[ssh_keys][]",
		"params[volumes][]",
		"params[read_only_config]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("username", request.Username)
	values.Add("params[password]", request.Password)
	values.Add("params[read_only_config]", strconv.Itoa(shtypes.BoolToInt(request.ReadOnlyConfig)))

	for _, c := range request.Containers {
		values.Add("params[containers][]", c)
	}

	for _, k := range request.SSHKeys {
		values.Add("params[ssh_keys][]", k)
	}

	for _, k := range request.Volumes {
		values.Add("params[volumes][]", k)
	}

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
