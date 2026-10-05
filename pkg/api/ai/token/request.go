package token

// Capabilities a token can be granted. Each unlocks one family of AI
// platform endpoints; a request outside the token's capabilities is
// refused by the AI platform with HTTP 403.
const (
	// CapabilityInference unlocks /v1/chat/completions.
	CapabilityInference = "inference"
	// CapabilityEmbedding unlocks /v1/embeddings.
	CapabilityEmbedding = "embedding"
	// CapabilityRerank unlocks /v1/rerank.
	CapabilityRerank = "rerank"
)

type (
	// AddRequest represents a request to issue a new AI token.
	AddRequest struct {
		// Label names the token. Required, 1–255 characters.
		Label string `json:"label"`

		// Capabilities the token is granted. Required, at least one
		// of the Capability* constants.
		Capabilities []string `json:"capabilities"`

		// AllowedIPs restricts where the token may be used from, as
		// bare IPs or CIDRs. Empty means usable from anywhere.
		AllowedIPs []string `json:"allowed_ips"`
	}

	// GetRequest represents a request to get a specific AI token.
	GetRequest struct {
		KeyID string `json:"key_id"`
	}

	// UpdateRequest represents a request to change an AI token.
	//
	// Every field other than KeyID is optional, and an unset field is
	// left as it is. An empty Label or Capabilities is unset; a token
	// must keep at least one capability, and the API rejects an empty
	// list.
	//
	// AllowedIPs is the exception, because clearing it is meaningful:
	// a nil AllowedIPs leaves the allowlist alone, and a non-nil empty
	// slice ([]string{}) removes every restriction.
	UpdateRequest struct {
		KeyID        string   `json:"key_id"`
		Label        string   `json:"label"`
		Capabilities []string `json:"capabilities"`
		AllowedIPs   []string `json:"allowed_ips"`
	}

	// DeleteRequest represents a request to revoke a specific AI token.
	DeleteRequest struct {
		KeyID string `json:"key_id"`
	}
)
