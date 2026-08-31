package api

import "net/http"

// SetTransport replaces the HTTP transport the client uses.
//
// This is the seam for anything that needs to sit between the SDK and
// the network: an outbound proxy, instrumentation, a caching layer, or
// a recorder for building test fixtures. Passing nil restores the
// default.
//
// The transport sees requests exactly as they go out, including the
// apikey query parameter, so anything logging from here must redact it —
// see [models.RedactURL].
func SetTransport(rt http.RoundTripper) ClientOpt {
	return func(c *Client) error {
		if rt == nil {
			rt = http.DefaultTransport
		}
		c.client.Transport = rt
		return nil
	}
}
