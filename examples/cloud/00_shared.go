package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sitehostnz/gosh/internal/recorder"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	cloudserver "github.com/sitehostnz/gosh/pkg/api/cloud/server"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
	stackimage "github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
	"github.com/sitehostnz/gosh/pkg/api/job"
	"github.com/sitehostnz/gosh/pkg/api/server"
)

// throttle spaces out calls. The API allows 10 requests per second by
// default, per reseller rather than per key; this is deliberately well
// under it.
const throttle = 1200 * time.Millisecond

// clients bundles what the journey drives.
type clients struct {
	server *server.Client
	cloud  *cloudserver.Client
	stack  *stack.Client
	image  *stackimage.Client
	db     *db.Client
	dbUser *dbuser.Client
	sshers *sshuser.Client
	job    *job.Client
}

// config holds the journey's inputs, all env-overridable.
type config struct {
	location string
	product  string
}

func newConfig() config {
	return config{
		location: envOr("SH_LOCATION", server.LocationAKLNCT),
		product:  envOr("SH_PRODUCT", "CLDCON4-P"),
	}
}

// state is what one step hands to the next.
type state struct {
	cfg config

	// name is the container server, either provisioned here or named by
	// SH_SERVER for a single-step run.
	name string
}

// newClients builds the API clients, optionally recording every call.
func newClients() (clients, error) {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return clients{}, fmt.Errorf("SH_API_KEY and SH_CLIENT_ID required")
	}

	var opts []api.ClientOpt

	// SH_BASE_URL is read here rather than merely documented. This is
	// the third time this exact fault has been raised in this
	// repository: an unrecognised variable that is silently ignored
	// means someone pointing a journey at a sandbox runs it against
	// production instead, and finds out afterwards.
	if base := os.Getenv("SH_BASE_URL"); base != "" {
		opts = append(opts, api.SetBaseURL(base))
	}

	if dir := os.Getenv("SH_RECORD_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return clients{}, fmt.Errorf("SH_RECORD_DIR: %w", err)
		}
		log.Printf("  recording every call to %s", dir)
		opts = append(opts, api.SetTransport(recorder.New(dir, nil)))
	}

	c, err := api.New(apiKey, clientID, opts...)
	if err != nil {
		return clients{}, fmt.Errorf("api.New: %w", err)
	}
	return clients{
		server: server.New(c),
		cloud:  cloudserver.New(c),
		stack:  stack.New(c),
		image:  stackimage.New(c),
		db:     db.New(c),
		dbUser: dbuser.New(c),
		sshers: sshuser.New(c),
		job:    job.New(c),
	}, nil
}

// envOr returns the environment variable or a fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// stepf logs a step heading.
func stepf(format string, args ...any) {
	log.Printf("── "+format, args...)
}
