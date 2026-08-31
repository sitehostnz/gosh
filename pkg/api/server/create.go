package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Create provisions a server via "server/provision.json".
//
// # Params were previously ignored
//
// Until August 2026 this method hardcoded params[ipv4]=auto and dropped
// every other field of [ParamsOptions] on the floor — IPv4, IPv6, Name,
// ContactID, Backup and SendEmail were all silently discarded, while
// the documentation described three IP-allocation paths of which only
// one was reachable. They are honoured now. Callers that relied on the
// old behaviour see no change: an empty Params.IPv4 still requests
// automatic allocation.
//
// # Addresses
//
// Leave Params.IPv4 empty, or set it to []string{"auto"}, to have the
// platform allocate one. Pass explicit addresses only if they are
// already allocated to the calling client; see [ParamsOptions].
//
// The two paths go on the wire differently, and the difference is not
// cosmetic. Automatic allocation is params[ipv4]=auto, a scalar;
// explicit addresses are params[ipv4][], repeated. Sending "auto" in
// the bracket form is rejected with
//
//	Error: The ip address is invalid, please specify a valid ip address
//
// which names an address the caller never supplied. This method sends
// the right form for each, so callers do not need to know.
//
// Array fields go on the wire in the bracket form the API expects
// (params[ipv4][], params[ipv6][], params[ssh_keys][]).
//
// # The name is not the label
//
// Params.Name proposes a name; the platform may not use it verbatim,
// and when it is left empty the name is derived from Label by
// truncating it and appending a digit on collision. Read
// CreateResponse.Return.Name and use that for every later call — see
// the package documentation.
//
// # SSH keys
//
// Params.SSHKeys takes public key *content*, not the ids returned by
// the ssh/key endpoints, and the keys must already be registered on the
// account or the provision is rejected. Both steps are required.
func (s *Client) Create(ctx context.Context, opts CreateRequest) (response CreateResponse, err error) {
	u := "server/provision.json"

	// Order matters: net.Encode emits keys in this order, and repeated
	// keys emit all their values.
	keys := []string{
		"client_id",
		"label",
		"location",
		"product_code",
		"image",
		"params[name]",
		"params[ipv4]",
		"params[ipv4][]",
		"params[ipv6][]",
		"params[ssh_keys][]",
		"params[contact_id]",
		"params[backup]",
		"params[send_email]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", opts.Label)
	values.Add("location", opts.Location)
	values.Add("product_code", opts.ProductCode)
	values.Add("image", opts.Image)

	if opts.Params.Name != "" {
		values.Add("params[name]", opts.Params.Name)
	}

	// An empty IPv4 means "allocate one", which is what the vast
	// majority of callers want and what this method used to hardcode.
	//
	// "auto" goes on the wire as a SCALAR, params[ipv4], not as
	// params[ipv4][]. The bracket form is rejected:
	//
	//	Error: The ip address is invalid, please specify a valid
	//	       ip address
	//
	// Explicit pre-allocated addresses use the bracket form instead,
	// since more than one may be passed. The two are not
	// interchangeable: sending "auto" in bracket form is rejected with
	// the message above, which names an address the caller never
	// supplied and so does not point at the real problem.
	ipv4 := opts.Params.IPv4
	if len(ipv4) == 0 {
		ipv4 = []string{"auto"}
	}
	if len(ipv4) == 1 && ipv4[0] == "auto" {
		values.Add("params[ipv4]", "auto")
	} else {
		for _, addr := range ipv4 {
			values.Add("params[ipv4][]", addr)
		}
	}
	for _, addr := range opts.Params.IPv6 {
		values.Add("params[ipv6][]", addr)
	}
	for _, key := range opts.Params.SSHKeys {
		values.Add("params[ssh_keys][]", key)
	}

	if opts.Params.ContactID != "" {
		values.Add("params[contact_id]", opts.Params.ContactID)
	}
	if opts.Params.Backup != "" {
		values.Add("params[backup]", opts.Params.Backup)
	}
	if opts.Params.SendEmail != "" {
		values.Add("params[send_email]", opts.Params.SendEmail)
	}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
