package srs

type (
	// ContactOptions identifies a contact by ID.
	ContactOptions struct {
		ContactID string `url:"contact_id"`
	}

	// SearchContactsOptions filters the contact list. At least one
	// of Name, Email, or RegistrantName must be set.
	SearchContactsOptions struct {
		Name           string `url:"query[name],omitempty"`
		Email          string `url:"query[email],omitempty"`
		RegistrantName string `url:"query[registrant_name],omitempty"`
		Offset         int    `url:"offsets[offset],omitempty"`
		Limit          int    `url:"offsets[limit],omitempty"`
	}
)
