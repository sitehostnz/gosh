package server

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// AddIP adds an IP address to a server via "server/add_ip.json".
// Set exactly one of opt.IP (a real address) or opt.IPVersion
// (4 or 6 to auto-allocate). See [AddIPOptions] for the rationale.
func (s *Client) AddIP(ctx context.Context, opt AddIPOptions) (response IPJobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("server.AddIP: Name is required")
	}
	param, err := addIPParam(opt)
	if err != nil {
		return response, err
	}

	values := url.Values{}
	values.Set("name", opt.Name)
	values.Set("param", param)

	req, err := s.client.NewRequest("POST", "server/add_ip.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

// addIPParam resolves the wire-level `param` value from AddIPOptions:
// the explicit address if IP is set, the family number if IPVersion
// is set. Exactly one must be supplied; both-set or both-unset is
// an error.
func addIPParam(opt AddIPOptions) (string, error) {
	hasIP := opt.IP != ""
	hasVersion := opt.IPVersion != 0
	switch {
	case hasIP && hasVersion:
		return "", fmt.Errorf("server.AddIP: set exactly one of IP or IPVersion (both supplied)")
	case !hasIP && !hasVersion:
		return "", fmt.Errorf("server.AddIP: set exactly one of IP or IPVersion (neither supplied)")
	case hasIP:
		return opt.IP, nil
	default:
		if opt.IPVersion != 4 && opt.IPVersion != 6 {
			return "", fmt.Errorf("server.AddIP: IPVersion must be 4 or 6, got %d", opt.IPVersion)
		}
		return strconv.Itoa(opt.IPVersion), nil
	}
}
