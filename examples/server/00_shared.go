package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/internal/recorder"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/job"
	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/api/server/firewall"
	"github.com/sitehostnz/gosh/pkg/api/server/firewall/securitygroups"
	"github.com/sitehostnz/gosh/pkg/api/server/snapshot"
	sshkey "github.com/sitehostnz/gosh/pkg/api/ssh/key"
	"github.com/sitehostnz/gosh/pkg/models"
)

const (
	// throttle spaces out calls. The API applies a per-second request
	// limit and rejects bursts with HTTP 500. This is deliberately far
	// more conservative than the limit: the journey is not in a hurry,
	// and the alternative is retry logic in every step.
	throttle = 1500 * time.Millisecond

	// jobTimeout bounds a single job. Provisioning is the slow one;
	// address moves complete in seconds.
	jobTimeout = 15 * time.Minute
)

// clients bundles every API client the journey needs, so step
// signatures stay short.
type clients struct {
	server *server.Client
	job    *job.Client
	fw     *firewall.Client
	sg     *securitygroups.Client
	key    *sshkey.Client
	snap   *snapshot.Client
}

// config holds the journey's inputs. Everything is env-overridable so
// no account detail is ever hardcoded.
type config struct {
	location string
	product  string
	image    string
	distro   string

	// labelA and labelB are the labels for the pair. The platform
	// derives the actual server names from these; see state.
	labelA string
	labelB string
}

func newConfig() config {
	return config{
		location: envOr("SH_LOCATION", server.LocationAKLNCT),
		product:  envOr("SH_PRODUCT", "LHPVS1"),
		// Image is discovered by step 20 rather than defaulted:
		// high-performance codes carry a build date and change when
		// images are rebuilt, so a literal here would rot.
		image:  os.Getenv("SH_IMAGE"),
		distro: envOr("SH_DISTRO", "ubuntu-noble"),
		labelA: envOr("SH_LABEL_A", "gosh-journey-a"),
		labelB: envOr("SH_LABEL_B", "gosh-journey-b"),
	}
}

// state is what one step hands to the next.
//
// When the whole journey runs in one process this is passed along in
// memory. When a single step is run on its own, the pieces it needs are
// read from the environment instead — see resolveServers.
type state struct {
	cfg config

	// keyID and publicKey identify the SSH key registered by step 10.
	keyID     string
	publicKey string
	// privateKey is held only in memory, never written to disk. Step 60
	// uses it to reach the guests.
	privateKey []byte

	// nameA and nameB are the names the platform assigned. These are
	// NOT the labels: the platform truncates the label and appends a
	// digit on collision.
	// subject is the server the read-only inventory step looks at. It
	// is deliberately separate from nameA/nameB, which the write steps
	// use: a read-only step must never redirect what a later write step
	// acts on.
	subject string

	nameA string
	nameB string

	// ipA and ipB are the primary addresses as they were before any
	// swap.
	ipA models.IP
	ipB models.IP

	// productType is the family reported by the API, which decides
	// whether some constraints apply.
	productType string

	// created lists every server this process provisioned, so step 90
	// can clean up even if an earlier step failed.
	created []string
}

// newClients builds the API clients from the environment.
func newClients() (clients, error) {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return clients{}, fmt.Errorf("SH_API_KEY and SH_CLIENT_ID required — see README.md for how to create a key")
	}
	var opts []api.ClientOpt

	// SH_BASE_URL is read here rather than merely documented. An
	// unrecognised variable that is silently ignored is the worst
	// outcome available: someone pointing this journey at a sandbox
	// before letting it loose would have provisioned, swapped and
	// deleted two billable servers on production instead.
	if base := os.Getenv("SH_BASE_URL"); base != "" {
		opts = append(opts, api.SetBaseURL(base))
	}

	// SH_RECORD_DIR turns on request/response recording. Every call and,
	// more usefully, every rejection is written there as JSON, so
	// fixtures can be derived from what the API actually did rather than
	// from what we assumed. Rejections cost nothing to collect: nothing
	// is provisioned and nothing has a side effect.
	if dir := os.Getenv("SH_RECORD_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return clients{}, fmt.Errorf("SH_RECORD_DIR: %w", err)
		}
		log.Printf("  recording API calls to %s", dir)
		opts = append(opts, api.SetTransport(recorder.New(dir, nil)))
	}

	c, err := api.New(apiKey, clientID, opts...)
	if err != nil {
		return clients{}, fmt.Errorf("api.New: %w", err)
	}

	// Say where this run is pointed, always. A journey that mutates
	// real infrastructure should never leave that implicit.
	log.Printf("  API base: %s", c.BaseURL)
	return clients{
		server: server.New(c),
		job:    job.New(c),
		fw:     firewall.New(c),
		sg:     securitygroups.New(c),
		key:    sshkey.New(c),
		snap:   snapshot.New(c),
	}, nil
}

// resolveServers fills in server names for a step run on its own,
// taking them from SH_SERVER_A / SH_SERVER_B.
func (s *state) resolveServers() error {
	if s.nameA == "" {
		s.nameA = os.Getenv("SH_SERVER_A")
	}
	if s.nameB == "" {
		s.nameB = os.Getenv("SH_SERVER_B")
	}
	if s.nameA == "" {
		return fmt.Errorf("no server to act on: run the whole journey, or set SH_SERVER_A (and SH_SERVER_B where a pair is needed)")
	}
	return nil
}

// deletable lists what this run should clean up.
//
// It returns only what this process provisioned. That restriction is
// the whole point, and it is worth stating why, because the convenient
// version of this function is dangerous.
//
// The convenient version falls back to SH_SERVER_A / SH_SERVER_B when
// nothing was recorded, so that a standalone `delete` has something to
// act on. But runJourney calls cleanup unconditionally, including when
// the tour failed — which is correct, since that is exactly when a
// freshly provisioned server would otherwise be leaked. Put those two
// together and a journey that fails before stepProvision records
// anything, at ssh/key.Create, or a preflight refusal, or a rate-limit
// burst, will find st.created empty and delete whatever SH_SERVER_A
// and SH_SERVER_B happen to name. Those are the variables the README
// tells you to export for standalone runs, so having them set
// alongside SH_EXAMPLE_ALLOW_PROVISION=1 is the normal working state,
// not an exotic one. Deletion is forced and not recoverable.
//
// Naming a server for a read-only step must never be reinterpreted as
// consent to destroy it. Deleting something this process did not create
// therefore needs its own opt-in, SH_DELETE_SERVERS, which exists for
// no other purpose and cannot be set by accident.
func (s *state) deletable() []string {
	if len(s.created) > 0 {
		return s.created
	}
	names := make([]string, 0, 2)
	for _, n := range strings.Split(os.Getenv("SH_DELETE_SERVERS"), ",") {
		if n = strings.TrimSpace(n); n != "" {
			names = append(names, n)
		}
	}
	if len(names) > 0 {
		log.Printf("  SH_DELETE_SERVERS names %d server(s) this process did not create: %s",
			len(names), strings.Join(names, ", "))
	}
	return names
}

// requirePair readies the state for a two-server step.
//
// Anything the journey would have recorded is read back from the API
// when it is missing, so a step can be run on its own — dropped into
// the middle of a journey someone else started — rather than only as
// part of a single process. That is the normal case for an agent
// picking up work, so it is the case this supports.
func (s *state) requirePair(ctx context.Context, c clients, step string) error {
	if err := s.resolveServers(); err != nil {
		return err
	}
	if s.nameB == "" {
		return fmt.Errorf("%s needs a pair; set SH_SERVER_B or run the whole journey", step)
	}
	return s.ensureAddresses(ctx, c)
}

// ensureAddresses fills in the current primary addresses, product
// family and login account for whatever servers are known, reading them
// from the API if this process did not provision them.
func (s *state) ensureAddresses(ctx context.Context, c clients) error {
	if err := s.fillAddress(ctx, c, &s.ipA, s.nameA); err != nil {
		return err
	}
	if err := s.fillAddress(ctx, c, &s.ipB, s.nameB); err != nil {
		return err
	}
	return s.fillPlatform(ctx, c)
}

// fillAddress reads one server's primary address if it is not known.
func (s *state) fillAddress(ctx context.Context, c clients, into *models.IP, name string) error {
	if into.IPAddr != "" || name == "" {
		return nil
	}
	time.Sleep(throttle)
	ip, err := primaryIPv4(ctx, c.server, name)
	if err != nil {
		return err
	}
	*into = ip
	return nil
}

// fillPlatform reads the product family and resolves the login account
// if this process did not provision the servers itself.
func (s *state) fillPlatform(ctx context.Context, c clients) error {
	if s.productType != "" {
		return nil
	}
	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: s.nameA})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", s.nameA, err)
	}
	s.productType = got.Server.ProductType
	if journeyLoginUser == "" {
		if user, ok := server.LoginUserFor(got.Server.ProductType, got.Server.Distro); ok {
			journeyLoginUser = user
			log.Printf("  resolved login account %q from %s/%s", user, got.Server.ProductType, got.Server.Distro)
		}
	}
	return nil
}

// waitJob polls a job to a terminal state. A zero ID means the call was
// synchronous and there is nothing to wait for.
func waitJob(ctx context.Context, j *job.Client, jb models.Job) error {
	if jb.ID == 0 {
		return nil
	}
	deadline := time.Now().Add(jobTimeout)
	for {
		time.Sleep(throttle * 2)
		resp, err := j.Get(ctx, job.GetRequest{ID: jb.ID, Type: jb.Type})
		if err != nil {
			// A poll can fail on the per-second limit; that is not the
			// job failing. Keep trying until the deadline.
			if time.Now().After(deadline) {
				return fmt.Errorf("job %d: %w", jb.ID, err)
			}
			continue
		}
		switch resp.Return.State {
		case "Completed":
			return nil
		case "Failed":
			return fmt.Errorf("job %d failed: %s", jb.ID, resp.Return.Message)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("job %d: still %s after %s", jb.ID, resp.Return.State, jobTimeout)
		}
	}
}

// primaryIPv4 returns a server's primary IPv4 address.
func primaryIPv4(ctx context.Context, s *server.Client, name string) (models.IP, error) {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return models.IP{}, fmt.Errorf("Get(%s): %w", name, err)
	}
	if len(resp.Server.Ips) == 0 {
		return models.IP{}, fmt.Errorf("Get(%s): server has no addresses", name)
	}
	for _, ip := range resp.Server.Ips {
		if ip.Primary && ip.AddrFamily == 4 {
			return ip, nil
		}
	}
	return models.IP{}, fmt.Errorf("Get(%s): no primary IPv4 among %d address(es)", name, len(resp.Server.Ips))
}

// ipCount reports how many addresses a server currently holds.
func ipCount(ctx context.Context, s *server.Client, name string) (int, error) {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return 0, fmt.Errorf("Get(%s): %w", name, err)
	}
	return len(resp.Server.Ips), nil
}

// assertHolds fails unless the named server currently holds addr.
func assertHolds(ctx context.Context, s *server.Client, name, addr string) error {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", name, err)
	}
	for _, ip := range resp.Server.Ips {
		if ip.IPAddr == addr {
			return nil
		}
	}
	return fmt.Errorf("%s does not hold %s (holds %d address(es))", name, addr, len(resp.Server.Ips))
}

// subnet labels an address's /24 (IPv4) or /64 (IPv6), for readable log
// lines only. Assertions compare NetworkID, which is the platform's own
// identity for an address space.
func subnet(addr string) string {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "unparseable"
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

// envOr returns the environment variable or a fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// step logs a step heading so a journey run is readable.
func stepf(format string, args ...any) {
	log.Printf("── "+format, args...)
}
