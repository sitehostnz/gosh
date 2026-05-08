package stack

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates a new cloud stack.
//
// # Gotchas (validated live, May 2026)
//
//  1. **Compose body must set `nz.sitehost.container.label` to a
//     valid FQDN** for any container with type=www or
//     type=application. The API rejects with "Unable to add stack,
//     the hostname is invalid." if this label's value isn't a
//     hostname-shaped string. The Label parameter on AddRequest
//     is *not* the field being validated here — the check is on
//     the compose body, not the API param. Set the same FQDN on
//     both for consistency.
//
//  2. **Stack Name must come from cloud.stack.GenerateName.**
//     The API rejects custom-shaped names with the same generic
//     "hostname is invalid" message; only platform-generated
//     "cc<hex>" names are accepted.
//
//  3. **Compose image references need an explicit version tag.**
//     `image: registry-clients.sitehost.co.nz/g_<id>/<code>` is
//     rejected with "There was no image version provided."; you
//     must include the `:1.0-<build_id>` tag from
//     cloud.image.version.list_all (or use the WaitForBuild
//     helper which surfaces it).
//
//  4. **Per-CCS write-time resource gate.** When the target CCS
//     is at capacity the API returns "the number of new images
//     required exceeds the number of available images on this
//     server." The read-side fields on cloud.server.List
//     (images_used / images_remaining) don't reliably reflect
//     the live cap; provision a fresh CCS or free a slot.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/stack/add.json"
	keys := []string{
		"client_id",
		"server",
		"name",
		"label",
		"enable_ssl",
		"docker_compose",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("enable_ssl", strconv.Itoa(request.EnableSSL))
	values.Add("docker_compose", request.DockerCompose)
	values.Add("label", request.Label)
	values.Add("name", request.Name)

	var vars string
	for _, envVar := range request.EnvironmentVariables {
		vars += fmt.Sprintf("  %s: %s\n", envVar.Name, envVar.Content)
	}

	if vars != "" {
		values.Add("environments["+request.Name+".env]", fmt.Sprintf("vars: \n%s", vars))
		keys = append(keys, "environments["+request.Name+".env]")
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
