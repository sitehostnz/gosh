package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db/grant"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack/integrated"
)

// stepDatabase creates a database, a user and grants, then proves the
// credentials work from inside the container.
//
// # MySQLHost is a stack name, not a hostname
//
// Despite the field name, cloud/db/add.json wants the name of a
// database stack on the same container server — resolvable only inside
// that server's Docker network. A fresh container server has no
// database stack at all, so this step deploys one before it can create
// anything in it. That is the reason this step is the longest in the
// journey.
//
// # What proves it, and what does not
//
// cloud/db/get.json returning the database, and cloud/db/user/list.json
// listing the user, are the control plane repeating what it was told.
// Neither can distinguish working credentials from a row in a table:
// a grant written against the wrong host, or a user created without the
// password taking effect, reports identically.
//
// So this step uses the mysql client, over the SSH account from step
// 40, to connect to the database stack as the user it created. It
// names the database it is in, creates a table, writes a row and reads
// it back — none of which goes through the API.
//
// Then it reduces the user to select-only with grant.Update and
// requires the database to refuse a write while still serving a read.
// That is the part a control-plane check cannot approach: a grant list
// stored in a table says nothing about what the user can do, and the
// only way to establish that grants are enforced is to lose one and
// notice.
//
// If the database stack never appears the step skips, loudly, and says
// it did not run. It does not fall back to asking the API whether the
// database exists — a check that quietly downgrades to one that cannot
// fail is worse than no check.
//
// # Consecutive Adds against one container collide
//
// cloud.db.Add holds a follow-on lock on its Container past the point
// its own job reports Completed, so a second Add against the same
// container is rejected with "There is already a job operating on the
// container." This step creates one database and does not hit it; see
// the note on db.Add if you are writing something that creates several.
func stepDatabase(ctx context.Context, c clients, st *state) error {
	name, err := subject(st)
	if err != nil {
		return err
	}
	if st.stackName == "" {
		return fmt.Errorf("no website container in state: run step 30 (stack) first")
	}
	if st.sshUser == "" {
		return fmt.Errorf("no ssh account in state: run step 40 (sshuser) first")
	}
	if err := ensureAddress(ctx, c, st, name); err != nil {
		return err
	}

	// Randomised per run. A fixed name collides with the journey's own
	// previous run — "Unable to create database, a database by this
	// name already exists" — and the database outlives a run that
	// failed after creating it.
	//
	// It is held locally until the create succeeds. Recording it up
	// front made the teardown try to delete a database that was never
	// created, with no host to delete it from: "The mysql_host
	// parameter is missing."
	database := "gosh_" + randomSuffix()
	if err := ensureDBStack(ctx, c, st, name); err != nil {
		return err
	}
	if err := createDatabase(ctx, c, st, name, database); err != nil {
		return err
	}
	if err := createUserAndGrants(ctx, c, st, name); err != nil {
		return err
	}
	if err := proveCredentials(st); err != nil {
		return err
	}
	if err := tightenGrants(ctx, c, st, name); err != nil {
		return err
	}
	return proveGrantsTookEffect(st)
}

// ensureDBStack makes sure the server has a database stack, and records
// which one.
//
// # MySQLHost is a stack name, and there is an endpoint for it
//
// cloud/db/add.json wants the name of a database stack on the same
// container server, resolvable inside its Docker network. Those stacks
// have their own endpoints — cloud/stack/integrated — and that is the
// only supported way to see or create one.
//
// It is worth saying plainly what the alternatives look like, because
// they are what you find first. cloud/stack/add.json refuses those
// names as reserved, listing them in the error, which reads as "you
// cannot create these" rather than "use the other endpoint". Deploying
// the same image under an unreserved name gets past that and then
// collides on the port the image publishes. Prefix-matching
// cloud/stack/list_all.json for "mysql" or "mariadb" finds them on some
// servers and not others. All three are dead ends, and none of them
// says so.
//
// # Why it is created rather than waited for
//
// A freshly provisioned container server has no database stack. An
// earlier version of this step waited fifteen minutes for one to appear
// on the theory that the platform added it asynchronously; it does not,
// and the theory came from one server that happened to have one because
// somebody had added it.
func ensureDBStack(ctx context.Context, c clients, st *state, name string) error {
	if v := os.Getenv("SH_MYSQL_HOST"); v != "" {
		st.mysqlHost = v
		log.Printf("  using the database stack named by SH_MYSQL_HOST: %s", v)
		return nil
	}

	time.Sleep(throttle)
	existing, err := c.integrated.List(ctx, integrated.ListRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("stack/integrated.List: %w", err)
	}
	if len(existing.Return) > 0 {
		st.mysqlHost = existing.Return[0]
		log.Printf("✓ %s already has database stack(s) %v; using %q",
			name, existing.Return, st.mysqlHost)
		return nil
	}

	wanted, err := newestDBImage(ctx, c)
	if err != nil {
		return err
	}
	return createDBStack(ctx, c, st, name, wanted)
}

// createDBStack adds an integrated stack and confirms it is listed.
func createDBStack(ctx context.Context, c clients, st *state, name, wanted string) error {
	log.Printf("  no database stack on %s; creating %q", name, wanted)

	time.Sleep(throttle)
	added, err := c.integrated.Add(ctx, integrated.AddRequest{ServerName: name, Name: wanted})
	if err != nil {
		return fmt.Errorf("stack/integrated.Add(%s): %w", wanted, err)
	}
	if added.Return.ID == 0 {
		// Guarded rather than assumed: version 1.0 of this endpoint
		// reports a bare job_id string, and a response that decodes to
		// no job would otherwise sail past waitJob, which treats a zero
		// id as nothing to wait for.
		return fmt.Errorf("stack/integrated.Add(%s): the response carried no job to wait for", wanted)
	}
	if err := waitJob(ctx, c.job, added.Return.Job); err != nil {
		return fmt.Errorf("stack/integrated.Add(%s): %w", wanted, err)
	}

	// Read it back. The job completing says the platform finished; this
	// says the stack is there under the name cloud/db will be asked
	// for, which is the part the next call depends on.
	time.Sleep(throttle)
	after, err := c.integrated.List(ctx, integrated.ListRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("stack/integrated.List after Add: %w", err)
	}
	for _, got := range after.Return {
		if got == wanted {
			st.mysqlHost = wanted
			log.Printf("✓ created database stack %q on %s", wanted, name)
			return nil
		}
	}
	return fmt.Errorf("stack/integrated.Add(%s) reported success but the stack is not listed afterwards; got %v",
		wanted, after.Return)
}

// newestDBImage picks a database stack to create from the image
// catalogue.
//
// The name of an integrated stack is its image code, and which codes
// exist is a property of the platform rather than of this example — so
// it is read rather than written down. PHPMyAdmin carries the same
// integrated type and is not a database, so it is excluded by name;
// there is no label distinguishing the two.
//
// The choice is the last MariaDB or MySQL code the catalogue lists,
// which in practice is the newest. It is deliberately not a constant:
// the set changes, and a journey pinned to one version fails when it is
// retired rather than when the SDK breaks.
func newestDBImage(ctx context.Context, c clients) (string, error) {
	time.Sleep(throttle)
	images, err := c.image.List(ctx)
	if err != nil {
		return "", fmt.Errorf("stack/image.List: %w", err)
	}

	candidates := make([]string, 0, len(images.Return))
	for _, im := range images.Return {
		// Labels is an untyped map: the label set differs per image
		// type, so no single struct fits them all.
		if t, _ := im.Labels["nz.sitehost.image.type"].(string); t != "integrated" {
			continue
		}
		if im.Code == integrated.StackPMA {
			continue
		}
		candidates = append(candidates, im.Code)
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no integrated database images in the catalogue")
	}
	log.Printf("  integrated database images offered: %v", candidates)
	return candidates[len(candidates)-1], nil
}

// createDatabase creates the database on the resolved stack.
func createDatabase(ctx context.Context, c clients, st *state, name, database string) error {
	time.Sleep(throttle)
	added, err := c.db.Add(ctx, db.AddRequest{
		ServerName: name,
		MySQLHost:  st.mysqlHost,
		Database:   database,
		// The website container owns the database, which is what
		// associates the two in the control panel.
		Container: st.stackName,
	})
	if err != nil {
		return fmt.Errorf("db.Add(%s on %s): %w", database, st.mysqlHost, err)
	}
	if err := waitJob(ctx, c.job, added.Return.Job); err != nil {
		return fmt.Errorf("db.Add(%s on %s): %w", database, st.mysqlHost, err)
	}
	// Recorded only once it exists, so the teardown either has a host
	// and a database or has neither.
	st.database = database
	log.Printf("✓ created database %q on stack %q", database, st.mysqlHost)
	return nil
}

// createUserAndGrants creates the database user and its grants.
//
// The database itself was created by resolveDBHost: finding the
// database stack requires attempting an Add against it, so the
// successful attempt is the creation. Splitting them would mean either
// probing with a throwaway database or probing with something other
// than the call whose behaviour is being established.
func createUserAndGrants(ctx context.Context, c clients, st *state, name string) error {
	st.dbUser = generatedUser()
	password, err := generatedSecret()
	if err != nil {
		return err
	}
	st.dbPassword = password

	// Lowercase, and that is not a style choice: the comparison is
	// case-sensitive, so the uppercase spelling every MySQL manual uses
	// is rejected with "one or more of the supplied grants are
	// invalid" — which does not say which one, or why.
	//
	// Twelve of the nineteen the platform accepts, deliberately. If the
	// count supplied equals the size of the whole set, the platform
	// substitutes ALL PRIVILEGES rather than enumerating them, and this
	// step then could not assert on the grants it asked for. A subset
	// keeps the request and the result comparable.
	grants := []string{
		"select", "insert", "update", "delete",
		"create", "drop", "index", "alter",
		"create temporary tables", "lock tables",
		"create view", "show view",
	}
	time.Sleep(throttle)
	user, err := c.dbUser.Add(ctx, dbuser.AddRequest{
		ServerName: name,
		MySQLHost:  st.mysqlHost,
		Username:   st.dbUser,
		Password:   st.dbPassword,
		Database:   st.database,
		Grants:     grants,
	})
	if err != nil {
		return fmt.Errorf("dbUser.Add(%s): %w", st.dbUser, err)
	}
	if err := waitJob(ctx, c.job, user.Return.Job); err != nil {
		return fmt.Errorf("dbUser.Add(%s): %w", st.dbUser, err)
	}
	log.Printf("✓ created database user %q with %d grant(s)", st.dbUser, len(grants))

	return confirmDatabaseReadable(ctx, c, st, name)
}

// confirmDatabaseReadable checks the control plane reports what was
// asked for. This is the weaker half; proveCredentials is the check
// that can fail for the right reasons.
func confirmDatabaseReadable(ctx context.Context, c clients, st *state, name string) error {
	time.Sleep(throttle)
	got, err := c.db.Get(ctx, db.GetRequest{
		ServerName: name, MySQLHost: st.mysqlHost, Database: st.database,
	})
	if err != nil {
		return fmt.Errorf("db.Get(%s): %w", st.database, err)
	}
	if got.Database.DBName != st.database {
		return fmt.Errorf("db.Get: returned %q, want %q", got.Database.DBName, st.database)
	}
	log.Printf("  control plane reports database %q on %q", got.Database.DBName, got.Database.MySQLHost)
	return nil
}

// proveCredentials connects to the database from inside the host, as
// the user this step created.
//
// # Why the mysql client and not a web page
//
// An earlier version wrote a PHP page into the website container and
// fetched it, so that a real request exercised the whole chain. That
// cannot work here: a stack created through the API is not reachable
// through the host's public proxy (see step 30), so the page could only
// be fetched from inside anyway — and once inside, the client is the
// more direct instrument. It also removes PHP from the path, so a
// failure is about the database rather than about the image.
//
// The SSH jail ships /usr/bin/mysql and a ~/.my.cnf naming the server's
// database hosts, which is what makes this available at all.
//
// What it establishes, none of it through the API: the credentials
// authenticate, the database exists, the grants permit writing to it,
// and the row written comes back.
func proveCredentials(st *state) error {
	// SELECT DATABASE() confirms which database the session is actually
	// in — connecting successfully to the wrong one would otherwise
	// read as a pass. The write and read back exercise the grants
	// rather than only the login.
	sql := strings.Join([]string{
		"SELECT DATABASE();",
		"CREATE TABLE IF NOT EXISTS gosh_probe (note VARCHAR(64));",
		"INSERT INTO gosh_probe VALUES ('" + st.database + "-ok');",
		"SELECT note FROM gosh_probe;",
	}, " ")

	// The password reaches mysql through MYSQL_PWD rather than the
	// command line, so it does not appear in the host's process list.
	script := fmt.Sprintf("MYSQL_PWD=%s mysql -h %s -u %s -D %s -N -B -e %s",
		shellQuote(st.dbPassword), shellQuote(st.mysqlHost),
		shellQuote(st.dbUser), shellQuote(st.database), shellQuote(sql))

	// Retried, for the same reason proveLogin is: a shell on a
	// freshly created container SSH account has a window where the
	// session opens and the command returns nothing at all. That was
	// observed here as an empty result reported as "the session did not
	// report the database", which reads like a database fault and is
	// not one. A database that genuinely never answers still fails,
	// with whatever the last attempt said.
	deadline := time.Now().Add(probeSettle)
	want := st.database + "-ok"
	var last string
	for {
		out, err := sshRunAs(st.sshUser, st.ip, script)
		switch {
		case err != nil:
			last = fmt.Sprintf("mysql failed: %v (%s)", err, firstLine(out))
		case strings.TrimSpace(out) == "":
			last = "the session returned no output at all"
		case !strings.Contains(out, st.database):
			last = fmt.Sprintf("the session did not report the database; output: %s", firstLine(out))
		case !strings.Contains(out, want):
			last = fmt.Sprintf("wrote a row but did not read it back; output: %s", firstLine(out))
		default:
			log.Printf("✓ out of band: connected to %q as %q, created a table, wrote a row and read it back",
				st.mysqlHost, st.dbUser)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("could not prove the credentials from inside %s within %s: %s", st.ip, probeSettle, last)
		}
		time.Sleep(5 * time.Second)
	}
}

// shellQuote renders a string as a single-quoted shell word.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// tightenGrants reduces the user to read-only through
// cloud/db/grant/update.json.
//
// # Why update and not add
//
// cloud/db/user/add.json already writes the user's grant for this
// database, so cloud/db/grant/add.json against the same pair is a
// duplicate and fails on a unique constraint. Add is for granting an
// existing user access to a *different* database; Update is what
// changes the grants on one it already has. That distinction is not
// documented and the failure does not suggest it — the endpoint returns
// the driver's own integrity-constraint error, including the SQL it
// tried to run.
//
// # What this makes checkable
//
// A grant list written into a table proves nothing about what the user
// can do. Reducing the grants and then finding that writes are refused
// while reads still work is the only version of this check that can
// fail for the right reason.
func tightenGrants(ctx context.Context, c clients, st *state, name string) error {
	time.Sleep(throttle)
	resp, err := c.grants.Update(ctx, grant.UpdateRequest{
		ServerName: name,
		MySQLHost:  st.mysqlHost,
		Username:   st.dbUser,
		Database:   st.database,
		Grants:     []string{"select"},
	})
	if err != nil {
		return fmt.Errorf("grant.Update(%s): %w", st.dbUser, err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("grant.Update(%s): %w", st.dbUser, err)
	}
	log.Printf("✓ grant.Update reduced %q to select only", st.dbUser)
	return nil
}

// proveGrantsTookEffect checks the tightened grants are enforced by the
// database, not merely recorded by the API.
//
// Two assertions, and both are needed. A refused insert alone could
// mean the credentials stopped working altogether; a working select
// alone could mean nothing changed. Together they say the grant was
// applied and scoped.
func proveGrantsTookEffect(st *state) error {
	// The database is polled because grant.Update returns a job and the
	// change is applied asynchronously — asserting immediately reads
	// the state from before the update.
	deadline := time.Now().Add(probeSettle)
	var last string
	for {
		readOK, err := mysqlSucceeds(st, "SELECT 1;")
		if err != nil {
			last = err.Error()
		}
		writeOK, _ := mysqlSucceeds(st, "INSERT INTO gosh_probe VALUES ('should-be-refused');")

		if err == nil && readOK && !writeOK {
			log.Printf("✓ out of band: with select-only grants the database refuses writes and still serves reads")
			return nil
		}
		if readOK && writeOK {
			last = "the insert still succeeded"
		}
		if !readOK {
			last = "the select stopped working too"
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("select-only grants were not enforced within %s: %s", probeSettle, last)
		}
		time.Sleep(5 * time.Second)
	}
}

// mysqlSucceeds runs one statement as the journey's database user and
// reports whether it succeeded.
func mysqlSucceeds(st *state, sql string) (bool, error) {
	script := fmt.Sprintf("MYSQL_PWD=%s mysql -h %s -u %s -D %s -N -B -e %s",
		shellQuote(st.dbPassword), shellQuote(st.mysqlHost),
		shellQuote(st.dbUser), shellQuote(st.database), shellQuote(sql))
	out, err := sshRunAs(st.sshUser, st.ip, script)
	if err == nil {
		return true, nil
	}
	// A statement the database refuses is the answer, not a transport
	// failure. Anything that is not the server talking is surfaced.
	if strings.Contains(out, "ERROR") || strings.Contains(out, "denied") {
		return false, nil
	}
	return false, fmt.Errorf("running mysql on %s failed before the server answered: %w (%s)", st.ip, err, firstLine(out))
}
