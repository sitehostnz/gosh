// Package integrated wraps the `/cloud/stack/integrated` endpoints,
// which manage the database stacks a Cloud Container's databases live
// in.
//
// # Why this is separate from cloud/stack
//
// A database stack is a stack, but it is not one you compose. The names
// it can take are reserved — cloud/stack/add.json refuses them, listing
// them in the error — because they belong to these endpoints instead.
// The platform builds the compose file itself from the image's latest
// version, including the environment file that holds the root password
// and the loopback port bindings, none of which a caller can supply.
//
// Trying to do it through cloud/stack/add.json therefore fails twice
// over: with the correct name it is rejected as reserved, and with
// another name it collides on the port the image publishes. Neither
// error mentions that a dedicated endpoint exists.
//
// # Why it matters for cloud/db
//
// cloud/db/add.json takes a MySQLHost, and despite the name that is the
// name of a database stack on the same server — mariadb1108, mysql57 —
// resolvable only inside that server's Docker network. If no such stack
// exists the call fails with "a stack does not currently exist for that
// MySQL version". [Client.List] is how you find out which are there,
// and [Client.Add] is how you get one.
//
// Note that integrated stacks are not reliably present in
// cloud/stack/list_all.json, so looking for a database host by
// prefix-matching that listing is not sound. Use [Client.List].
package integrated
