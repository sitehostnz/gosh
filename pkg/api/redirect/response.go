package redirect

import (
	"encoding/json"

	"github.com/sitehostnz/gosh/pkg/models"
)

// ListRedirectsResponse is the return from /redirect/list_redirects.json.
//
// **Wire-shape quirk** (verified live, May 2026): when the account
// has redirects, "return" is a nested map keyed by domain → source
// URL → Rule:
//
//	{
//	  "example.kiwi.nz": {
//	    "a.example.kiwi.nz": {"destination": "https://...", "type": 301},
//	    "b.example.kiwi.nz/blog": {"destination": "https://...", "type": 302}
//	  }
//	}
//
// When the account has **zero** redirects, "return" is the JSON
// array `[]`, not the empty object `{}`. The docs only show the
// populated form, so consumers parsing the obvious map shape would
// fail on empty accounts.
//
// Custom UnmarshalJSON handles both: the empty array is decoded as
// an empty map; the populated map is decoded normally.
type ListRedirectsResponse struct {
	Return map[string]map[string]Rule `json:"return"`
	models.APIResponse
}

// UnmarshalJSON tolerates the empty-array form the API returns
// when the account has no redirect rules. See type comment.
func (r *ListRedirectsResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Return json.RawMessage `json:"return"`
		models.APIResponse
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.APIResponse = envelope.APIResponse

	raw := envelope.Return
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" {
		r.Return = map[string]map[string]Rule{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}
