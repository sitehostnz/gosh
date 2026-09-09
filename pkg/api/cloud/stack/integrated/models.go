package integrated

import (
	"github.com/sitehostnz/gosh/pkg/api"
)

type (
	// Client is a service to work with integrated (database) stacks.
	Client struct {
		client *api.Client
	}
)

// New is an initialisation function.
func New(c *api.Client) *Client {
	return &Client{
		client: c,
	}
}

// The stack names these endpoints accept.
//
// A database stack's name is its image code, and the set is fixed: the
// platform rejects anything else with "invalid label passed. Please
// specify one from the following", listing these. They are the same
// names cloud/stack/add.json refuses, for the same reason.
//
// Prefer discovering what a server can take from
// cloud/stack/image/list_all.json, filtering on the
// nz.sitehost.image.type label being "integrated" — these constants are
// for readability at call sites, not a substitute for the catalogue.
// PMA is in the set and is PHPMyAdmin rather than a database.
const (
	StackMariaDB1011 = "mariadb1011"
	StackMariaDB1108 = "mariadb1108"
	StackMySQL56     = "mysql56"
	StackMySQL57     = "mysql57"
	StackMySQL8      = "mysql8"
	StackMySQL84     = "mysql84"
	StackPMA         = "pma"
)
