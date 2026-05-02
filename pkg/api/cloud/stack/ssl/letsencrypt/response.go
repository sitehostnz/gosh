package letsencrypt

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// CertInfo is a single LE cert's metadata as returned by
	// list_all. Per-stack info is keyed by the stack name in the
	// outer Return map of ListResponse.
	//
	// String-typed boolean fields (Expired, IsMissing) reflect the
	// API's actual response — values arrive as strings ("0"/"1")
	// rather than typed bools.
	CertInfo struct {
		Issuer    string `json:"issuer"`
		NotBefore string `json:"not_before"`
		NotAfter  string `json:"not_after"`
		Serial    string `json:"serial"`
		Expired   string `json:"expired"`
		IsMissing string `json:"is_missing"`
	}

	// ListResponse represents the response from list_all. Return is
	// a map keyed by stack name to that stack's cert metadata. A
	// stack with no LE cert simply doesn't appear in the map.
	ListResponse struct {
		Return map[string]CertInfo `json:"return"`
		models.APIResponse
	}

	// JobResponse is the shared response shape for the asynchronous
	// write operations: Create, Delete, Renew, Revoke. Each queues
	// a scheduler job; the job id is returned for tracking.
	JobResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
