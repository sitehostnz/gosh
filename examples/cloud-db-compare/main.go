// Program cloud-db-compare creates a database + user on each of a
// CCS's MariaDB and MySQL stacks, prints a side-by-side comparison
// of the responses, and tears down.
//
// # How "which database server" is chosen
//
// A SiteHost Cloud Container Server (CCS) can run multiple
// database engines simultaneously, each as its own Docker stack.
// The stack names ARE the engine identifiers — typical examples:
//
//   mariadb1108  — MariaDB 11.8
//   mysql57      — MySQL 5.7
//   mysql8       — MySQL 8.x
//   postgres15   — PostgreSQL 15
//
// (See cloud.stack.image.list_all for the catalog of available
// engines + the exact stack code shapes.)
//
// When creating a database via cloud.db.Add, the `mysql_host`
// field selects which DB stack on the CCS the database lives in.
// It is NOT a hostname in the DNS sense — it's the stack name
// that's resolvable inside the CCS Docker network.
//
// This example discovers what DB stacks are present on the target
// CCS by listing stacks and matching well-known prefixes
// (mariadb*, mysql*, postgres*). For each match it provisions a
// throwaway database, then prints what each engine returned so
// you can see how the platform reports them differently.
//
// # Required env
//
//   SH_API_KEY     — API key
//   SH_CLIENT_ID   — client id
//   SH_CCS_NAME    — name of the Cloud Container Server with the
//                    database stacks deployed
//
// # Optional env
//
//   SH_CONTAINER_NAME — existing www-type container name on the
//                       CCS to use as the database's owning
//                       container (cloud.db.Add's `container`
//                       field). If unset, the example picks the
//                       first non-DB / non-infra stack it finds on
//                       the CCS. Provide explicitly if your CCS
//                       has multiple containers and you want to
//                       pin the association.
//   JOURNEY_KEEP=1  — leave the created databases + users in place
//                     (default: cleanup).
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api"
	cloudDB "github.com/sitehostnz/gosh/pkg/api/cloud/db"
	cloudDBUser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
	"github.com/sitehostnz/gosh/pkg/api/job"
)

// dbStackPrefixes — known database stack name prefixes. The
// matched prefix becomes the engine label in the comparison.
var dbStackPrefixes = []struct {
	prefix, engine string
}{
	{"mariadb", "MariaDB"},
	{"mysql", "MySQL"},
	{"postgres", "PostgreSQL"},
}

type dbResult struct {
	engine   string
	stackName string
	dbName    string
	dbUser    string
	dbPass    string
	dbInfo    string // dump of the cloud.db.Get response
	addErr    error
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("cloud-db-compare: %v", err)
	}
}

func run() error {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	ccsName := os.Getenv("SH_CCS_NAME")
	if apiKey == "" || clientID == "" || ccsName == "" {
		return fmt.Errorf("SH_API_KEY, SH_CLIENT_ID, SH_CCS_NAME required")
	}
	keep := os.Getenv("JOURNEY_KEEP") == "1"
	c, err := api.New(apiKey, clientID)
	if err != nil {
		return fmt.Errorf("api.New: %w", err)
	}
	ctx := context.Background()

	// 1. Discover stacks on the CCS — we need to know what DB
	//    engines are available and pick a non-DB container as the
	//    "owner" for cloud.db.Add's `container` field.
	log.Printf("==> listing stacks on %s", ccsName)
	stacks, err := stack.New(c).List(ctx, stack.ListRequest{ServerName: ccsName})
	if err != nil {
		return fmt.Errorf("cloud.stack.List: %w", err)
	}
	log.Printf("    %d stack(s)", len(stacks.Return.Stacks))

	dbStacks := []dbResult{}
	containerName := os.Getenv("SH_CONTAINER_NAME")
	for _, s := range stacks.Return.Stacks {
		matched := false
		for _, p := range dbStackPrefixes {
			if strings.HasPrefix(s.Name, p.prefix) {
				dbStacks = append(dbStacks, dbResult{engine: p.engine, stackName: s.Name})
				log.Printf("    [db]  %s (%s)", s.Name, p.engine)
				matched = true
				break
			}
		}
		if !matched && containerName == "" && s.Name != "infra" {
			containerName = s.Name // first non-DB / non-infra stack
		}
		if !matched {
			log.Printf("    [app] %s", s.Name)
		}
	}

	if len(dbStacks) < 2 {
		return fmt.Errorf("need at least 2 DB stacks on %s for a comparison; found %d",
			ccsName, len(dbStacks))
	}
	if containerName == "" {
		return fmt.Errorf("no non-DB / non-infra container on %s — set SH_CONTAINER_NAME explicitly", ccsName)
	}
	log.Printf("    using container=%s as the owning container", containerName)

	// Sort dbStacks for stable output.
	sort.Slice(dbStacks, func(i, j int) bool { return dbStacks[i].engine < dbStacks[j].engine })

	suffix := randHex(6)

	// Defer cleanup early so a partial run still tears down what
	// it created.
	defer func() {
		if keep {
			log.Printf("JOURNEY_KEEP=1 — leaving databases + users in place")
			return
		}
		log.Printf("==> cleanup")
		for _, d := range dbStacks {
			if d.dbName == "" {
				continue
			}
			cleanupDB(ctx, c, ccsName, d.stackName, d.dbName, d.dbUser)
		}
	}()

	// 2. For each DB stack, create database + user.
	//
	// Naming notes (verified live, May 2026):
	//
	//   - Database names accept underscores; tested goshdb_<engine>_<hex>
	//     successfully.
	//   - **Usernames are restricted** — the API rejects underscored
	//     names with "Please specify a valid username." Mirroring
	//     the cc<hex> stack-name shape (alphanumeric, ≤16 chars)
	//     works. Picked "g<short><6hex>" here.
	//
	// Between the two DB adds we sleep briefly to let the
	// container-level lock release. Without it the second
	// cloud.db.Add hits "Unable to create database. There is
	// already a job operating on the container." even though the
	// scheduler job from the first add reported Completed.
	for i := range dbStacks {
		if i > 0 {
			log.Printf("    sleeping 30s for container-level lock to release before next DB add")
			time.Sleep(30 * time.Second)
		}
		d := &dbStacks[i]
		shortEngine := strings.ToLower(d.engine[:1]) // m / y / p
		d.dbName = fmt.Sprintf("goshdb_%s_%s", strings.ToLower(d.engine), suffix)
		d.dbUser = fmt.Sprintf("g%s%s", shortEngine, suffix) // e.g. gm9f5508 (8 chars)
		d.dbPass = randHex(16)

		log.Printf("==> %s: cloud.db.Add(%s)", d.engine, d.dbName)
		addResp, err := cloudDB.New(c).Add(ctx, cloudDB.AddRequest{
			ServerName: ccsName,
			MySQLHost:  d.stackName,
			Database:   d.dbName,
			Container:  containerName,
		})
		if err != nil {
			d.addErr = err
			log.Printf("    error: %v", err)
			continue
		}
		if err := waitForJob(ctx, c, addResp.Return.ID, addResp.Return.Type, 2*time.Minute); err != nil {
			d.addErr = err
			log.Printf("    job: %v", err)
			continue
		}
		log.Printf("    ✓ database created")

		log.Printf("==> %s: cloud.db.user.Add(%s)", d.engine, d.dbUser)
		userResp, err := cloudDBUser.New(c).Add(ctx, cloudDBUser.AddRequest{
			ServerName: ccsName,
			MySQLHost:  d.stackName,
			Username:   d.dbUser,
			Password:   d.dbPass,
			Database:   d.dbName,
			Grants: []string{
				"select", "insert", "update", "delete",
				"create", "drop", "index", "alter",
				"create temporary tables", "lock tables",
				"create view", "show view",
			},
		})
		if err != nil {
			d.addErr = err
			log.Printf("    error: %v", err)
			continue
		}
		if err := waitForJob(ctx, c, userResp.Return.ID, userResp.Return.Type, 2*time.Minute); err != nil {
			d.addErr = err
			log.Printf("    job: %v", err)
			continue
		}
		log.Printf("    ✓ user + grants applied")

		// Read back to capture how each engine reports the DB.
		got, err := cloudDB.New(c).Get(ctx, cloudDB.GetRequest{
			ServerName: ccsName,
			MySQLHost:  d.stackName,
			Database:   d.dbName,
		})
		if err != nil {
			log.Printf("    Get: %v", err)
			continue
		}
		d.dbInfo = fmt.Sprintf("id=%s db_name=%s mysql_host=%s size=%v client_id=%s",
			got.Database.ID, got.Database.DBName, got.Database.MySQLHost,
			got.Database.Size, got.Database.ClientID)
	}

	// 3. Comparison output.
	log.Println()
	log.Println("============== comparison ==============")
	for _, d := range dbStacks {
		log.Printf("[%s]", d.engine)
		log.Printf("  stack:      %s", d.stackName)
		log.Printf("  database:   %s", d.dbName)
		log.Printf("  user:       %s", d.dbUser)
		log.Printf("  password:   %s (scrubbed in real output; printed here for the example)", d.dbPass[:4]+"…")
		if d.addErr != nil {
			log.Printf("  status:     ERR: %v", d.addErr)
		} else {
			log.Printf("  Get():      %s", d.dbInfo)
		}
		log.Println()
	}
	return nil
}

func cleanupDB(ctx context.Context, c *api.Client, ccsName, mysqlHost, dbName, dbUser string) {
	if dbUser != "" {
		_, err := cloudDBUser.New(c).Delete(ctx, cloudDBUser.DeleteRequest{
			ServerName: ccsName, MySQLHost: mysqlHost, Username: dbUser,
		})
		if err != nil {
			log.Printf("    [%s] user delete: %v", dbName, err)
		} else {
			log.Printf("    [%s] ✓ user deleted", dbName)
		}
	}
	resp, err := cloudDB.New(c).Delete(ctx, cloudDB.DeleteRequest{
		ServerName: ccsName, MySQLHost: mysqlHost, Database: dbName,
	})
	if err != nil {
		log.Printf("    [%s] db delete: %v", dbName, err)
		return
	}
	if resp.Return.ID > 0 {
		if err := waitForJob(ctx, c, resp.Return.ID, resp.Return.Type, 2*time.Minute); err != nil {
			log.Printf("    [%s] db delete job: %v", dbName, err)
			return
		}
	}
	log.Printf("    [%s] ✓ db deleted", dbName)
}

func randHex(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)[:n]
}

func waitForJob(ctx context.Context, c *api.Client, id int, jobType string, timeout time.Duration) error {
	jc := job.New(c)
	deadline := time.Now().Add(timeout)
	for {
		resp, err := jc.Get(ctx, job.GetRequest{ID: id, Type: jobType})
		if err != nil {
			return err
		}
		switch resp.Return.State {
		case "Completed":
			return nil
		case "Failed":
			return fmt.Errorf("job %d failed (state=%s)", id, resp.Return.State)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("job %d timed out (state=%s)", id, resp.Return.State)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}
