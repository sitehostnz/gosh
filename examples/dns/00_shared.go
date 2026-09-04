package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sitehostnz/gosh/internal/recorder"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/dns"
	"github.com/sitehostnz/gosh/pkg/api/dns/template"
)

// throttle spaces out calls. The API allows 10 requests per second per
// reseller — not per key — and this sits well under it.
const throttle = 1200 * time.Millisecond

// clients bundles what the journey drives.
type clients struct {
	dns      *dns.Client
	template *template.Client
}

// config holds the journey's inputs.
type config struct {
	// zone is the DNS zone this journey creates and destroys.
	//
	// # Why not a reserved name
	//
	// The obvious safe choice is something under .invalid, reserved by
	// RFC 2606 so it can never be registered. The API rejects it:
	//
	//	CreateZone: 400 Please specify a valid domain name.
	//
	// It validates the top-level domain, so .invalid, .test and
	// .example are all unavailable. That leaves a real TLD, and the
	// only way to stay safe is to make collision with something real
	// vanishingly unlikely — hence the random label below.
	//
	// Creating a zone is not registering a domain. It costs nothing and
	// has no effect unless something delegates to SiteHost's
	// nameservers, which nothing will for a name that does not exist.
	zone string
}

func newConfig() config {
	return config{zone: envOr("SH_ZONE", generatedZone())}
}

// generatedZone builds a name unlikely to collide with a real domain.
//
// Random rather than fixed, because a fixed name turns two concurrent
// runs into one run deleting the other's zone — and because a leftover
// from a failed run would then block every later one.
func generatedZone() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Randomness failing is not a reason to fall back to a
		// predictable name; that is the case this guards against.
		panic("dns example: cannot generate a zone name: " + err.Error())
	}
	return fmt.Sprintf("gosh-example-%s.co.nz", hex.EncodeToString(b[:]))
}

// state is what one step hands to the next.
type state struct {
	cfg config

	// created records whether this process made the zone.
	//
	// Cleanup deletes only what it created. A zone named in SH_ZONE is
	// never removed implicitly — the server journey had to learn that
	// rule the hard way, when naming a server for a read-only step
	// turned out to be read as consent to destroy it.
	created bool

	// recordID is the record added at step 30, carried to the steps
	// that change and remove it.
	recordID string

	// originalTemplate is the zone's template before step 40 changed
	// it, so the change can be put back.
	originalTemplate string
}

// newClients builds the API clients, optionally recording every call.
func newClients() (clients, error) {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return clients{}, fmt.Errorf("SH_API_KEY and SH_CLIENT_ID required")
	}

	var opts []api.ClientOpt
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
	log.Printf("  API base: %s", c.BaseURL)

	return clients{dns: dns.New(c), template: template.New(c)}, nil
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
