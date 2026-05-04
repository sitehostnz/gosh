package srs

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
)

// NameServerEntry is a single nameserver record to add to a domain
// via "srs/add_name_servers.json". Name is required; IPv4Addr and
// IPv6Addr are only required when registering glue records (the
// nameserver is itself a hostname inside the domain you're
// updating). For external nameservers (ns1.somewhere-else.com)
// leave IPv4Addr and IPv6Addr empty.
type NameServerEntry struct {
	Name     string
	IPv4Addr string
	IPv6Addr string
}

// AddNameServersOptions adds one or more nameservers to a domain.
// Domain and at least one entry in NameServers are required.
//
// The wire format is array-indexed:
//
//	nameservers[0][name]=ns1.example.com
//	nameservers[0][ipv4addr]=192.0.2.1     (optional, glue only)
//	nameservers[1][name]=ns2.example.com
//	...
//
// Add is additive — existing nameservers are preserved unless the
// registry replaces them. To replace the full set, you typically
// remove (registry-side; not currently exposed) and re-add.
type AddNameServersOptions struct {
	Domain      string
	NameServers []NameServerEntry
}

// AddNameServers attaches one or more nameservers to a domain via
// "srs/add_name_servers.json".
func (s *Client) AddNameServers(ctx context.Context, opt AddNameServersOptions) (response models.APIResponse, err error) {
	if opt.Domain == "" {
		return response, fmt.Errorf("srs.AddNameServers: Domain is required")
	}
	if len(opt.NameServers) == 0 {
		return response, fmt.Errorf("srs.AddNameServers: at least one NameServer is required")
	}
	values := url.Values{}
	values.Set("domain", opt.Domain)
	for i, ns := range opt.NameServers {
		if ns.Name == "" {
			return response, fmt.Errorf("srs.AddNameServers: NameServers[%d].Name is required", i)
		}
		idx := strconv.Itoa(i)
		values.Set("nameservers["+idx+"][name]", ns.Name)
		if ns.IPv4Addr != "" {
			values.Set("nameservers["+idx+"][ipv4addr]", ns.IPv4Addr)
		}
		if ns.IPv6Addr != "" {
			values.Set("nameservers["+idx+"][ipv6addr]", ns.IPv6Addr)
		}
	}
	req, err := s.client.NewRequest("POST", "srs/add_name_servers.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
