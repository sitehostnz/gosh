package srs

type (
	// WhoisOptions represents the parameters for a whois lookup. Domain
	// is required.
	WhoisOptions struct {
		Domain string `url:"domain"`
	}
)
