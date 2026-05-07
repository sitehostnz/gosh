package letsencrypt

type (
	// ListRequest identifies a stack on a CCS whose Let's Encrypt
	// certs to list. ServerName is the CCS name (the unprefixed
	// "server" parameter that cloud/stack/* uses, not "server_name").
	// StackName is the stack's name within that server. Containers
	// is an optional filter restricting the result to specific
	// containers within the stack.
	//
	// Both ServerName and StackName are required by the API; omitting
	// the stack name returns "The stack name is missing."
	ListRequest struct {
		ServerName string   `url:"server"`
		StackName  string   `url:"name"`
		Containers []string `url:"containers,omitempty"`
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
