package letsencrypt

type (
	// ListRequest identifies the cloud server whose stacks' LE certs
	// to enumerate. ServerName is the CCS name (the unprefixed
	// "server" parameter that cloud/stack/* uses, not "server_name").
	ListRequest struct {
		ServerName string `url:"server"`
	}

	// CreateRequest queues an LE-cert issuance for the named stack.
	// HTTP-01 validation runs against the stack's nginx-proxy
	// vhost; the stack must be reachable on port 80 at its label
	// hostname for the challenge to succeed.
	CreateRequest struct {
		ServerName string `url:"server"`
		Name       string `url:"name"`
	}

	// DeleteRequest removes the LE cert for the named stack.
	DeleteRequest struct {
		ServerName string `url:"server"`
		Name       string `url:"name"`
	}

	// RenewRequest forces an early renewal of the LE cert for the
	// named stack. Normally the companion auto-renews on schedule;
	// use this for out-of-band refresh.
	RenewRequest struct {
		ServerName string `url:"server"`
		Name       string `url:"name"`
	}

	// RevokeRequest revokes the LE cert for the named stack at the
	// CA. Distinct from Delete: revocation is a CA-side action that
	// invalidates the cert across the public PKI.
	RevokeRequest struct {
		ServerName string `url:"server"`
		Name       string `url:"name"`
	}
)
