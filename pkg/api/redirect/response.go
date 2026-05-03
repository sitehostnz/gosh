package redirect

import "github.com/sitehostnz/gosh/pkg/models"

// ListRedirectsResponse is the return from /redirect/list_redirects.json.
//
// The API returns a nested map: the outer key is the domain, the
// inner key is the source URL (host + optional path), and the value
// is the Rule (destination + status code).
//
//	{
//	  "example.kiwi.nz": {
//	    "a.example.kiwi.nz": {"destination": "https://...", "type": 301},
//	    "b.example.kiwi.nz/blog": {"destination": "https://...", "type": 302}
//	  }
//	}
//
// Modelled as map[domain]map[sourceURL]Rule rather than a flat list
// to match the wire shape; consumers wanting a flat slice can range
// over both levels themselves.
type ListRedirectsResponse struct {
	Return map[string]map[string]Rule `json:"return"`
	models.APIResponse
}
