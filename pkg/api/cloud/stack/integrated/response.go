package integrated

import (
	"github.com/sitehostnz/gosh/pkg/models"
)

type (
	// ListResponse is the response from
	// cloud/stack/integrated/list_all.json.
	//
	// The return is a bare array of stack names — ["mariadb1108",
	// "mysql57"] — not the object-with-data shape the other cloud
	// listings use, and it carries no pagination.
	ListResponse struct {
		Return []string `json:"return"`
		models.APIResponse
	}

	// AddResponse is the response from
	// cloud/stack/integrated/add.json.
	//
	// The job arrives in the same {"job":{"id":…,"type":…}} shape the
	// sibling stack endpoints use.
	//
	// Version 1.0 of this endpoint does not: it returns a bare
	// "job_id" as a JSON string. This type is written for 1.5, which is
	// the version this SDK targets — worth stating because the 1.0
	// source is what a search finds first, and a response type built
	// from it decodes 1.5 into an empty job. That is not a harmless
	// mistake: job.Get reads ID == 0 as "nothing to wait for", so the
	// caller skips the wait and treats an unbuilt stack as built. It
	// was written that way here first, and the wire caught it.
	AddResponse struct {
		Return struct {
			models.Job `json:"job"`
		} `json:"return"`
		models.APIResponse
	}
)
