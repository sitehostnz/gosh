package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/internal/recorder"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db/grant"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	cloudserver "github.com/sitehostnz/gosh/pkg/api/cloud/server"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
	stackimage "github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack/integrated"
	"github.com/sitehostnz/gosh/pkg/api/job"
	"github.com/sitehostnz/gosh/pkg/api/server"
	sshkey "github.com/sitehostnz/gosh/pkg/api/ssh/key"
	"github.com/sitehostnz/gosh/pkg/models"
	"golang.org/x/crypto/ssh"
)

// throttle spaces out calls. The API allows 10 requests per second by
// default, per reseller rather than per key; this is deliberately well
// under it.
const throttle = 1200 * time.Millisecond

// clients bundles what the journey drives.
type clients struct {
	server     *server.Client
	cloud      *cloudserver.Client
	stack      *stack.Client
	image      *stackimage.Client
	integrated *integrated.Client
	db         *db.Client
	dbUser     *dbuser.Client
	grants     *grant.Client
	sshers     *sshuser.Client
	sshkeys    *sshkey.Client
	job        *job.Client
}

// config holds the journey's inputs, all env-overridable.
type config struct {
	location string

	// product is the container product code. Empty means "resolve the
	// smallest one offered at this location", which is what step 10
	// does. Hardcoding a code was the previous behaviour and is exactly
	// what server/products.json exists to avoid — there are 21 CLDCON
	// codes at AKLNCT alone, and which are offered varies by location.
	product string

	// image is sent as the provision image code.
	//
	// It must be a code that server.ListImages actually returns. The
	// endpoint accepts an unknown code at request time — nothing is
	// rejected, a job id comes back — and the build then sits in
	// "Configuring server" for twenty minutes before failing, after
	// which the platform deletes the server it was building. That was
	// observed here with "ubuntu-noble.amd64", which does not exist;
	// the real code is "ubuntu-noble-pvh.amd64".
	//
	// So "the image parameter is not validated" is true only of the
	// request. Step 10 checks the code against the catalogue before
	// anything is provisioned, because a typo otherwise costs twenty
	// minutes and reports nothing useful.
	image string

	// label is the provision label. The platform derives the server
	// name from it, so the name this journey uses is the returned one.
	//
	// It is randomised per run, and the randomness is deliberately at
	// the front. The derived name is "ch-" followed by the label's
	// first nine characters — the label "gosh-cloud-journey" produced
	// "ch-gosh-clou" — so a suffix would be truncated away and every
	// run would ask for the same name. Reusing the name of a server
	// deleted moments earlier fails the build with "There was a problem
	// creating your server, please contact support", ninety seconds in
	// and with nothing to say why.
	label string
}

func newConfig() config {
	return config{
		location: envOr("SH_LOCATION", server.LocationAKLNCT),
		product:  os.Getenv("SH_PRODUCT"),
		image:    envOr("SH_IMAGE", "ubuntu-noble-pvh.amd64"),
		label:    envOr("SH_LABEL", generatedLabel()),
	}
}

// state is what one step hands to the next.
type state struct {
	cfg config

	// name is the container server, either provisioned here or named by
	// SH_SERVER for a single-step run.
	name string

	// ip is the container server's primary IPv4 address. Every
	// out-of-band check in this journey goes to it rather than to the
	// API.
	ip string

	// created holds the server names this process provisioned, and is
	// the only thing step 90 will delete. A server named by SH_SERVER is
	// deliberately absent: pointing the journey at something that
	// already exists must not make it deletable.
	created []string

	// publicKey is the journey key in authorized_keys form, handed to
	// cloud/ssh/user.Add. privateKey is its half, used to log in and
	// prove the user works.
	publicKey  string
	privateKey ed25519.PrivateKey

	// imageChecked records that the image code has been validated
	// against the catalogue, so a full journey does not repeat the call
	// in step 20 after step 10 has already made it.
	imageChecked bool

	// sshUser is the container SSH account created by step 40, and
	// sshKeyID the account key registered for it.
	sshUser  string
	sshKeyID string

	// stackName is the generated stack name, which is also the
	// container name and the compose service key. stackHost is the
	// wildcard hostname the stack answers on.
	stackName string
	stackHost string

	// database, dbUser and dbPassword are created by step 50, and
	// mysqlHost is the host the grant is scoped to.
	database   string
	dbUser     string
	dbPassword string
	mysqlHost  string
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
		server:     server.New(c),
		cloud:      cloudserver.New(c),
		stack:      stack.New(c),
		image:      stackimage.New(c),
		integrated: integrated.New(c),
		db:         db.New(c),
		dbUser:     dbuser.New(c),
		grants:     grant.New(c),
		sshers:     sshuser.New(c),
		sshkeys:    sshkey.New(c),
		job:        job.New(c),
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

// jobTimeout bounds a provision or delete.
//
// This is deliberately longer than the server journey's fifteen
// minutes. A failing container build was observed sitting in
// "Configuring server" for twenty minutes before the job reported
// Failed, so a fifteen-minute budget times out first and reports
// "still Running" — hiding the actual failure and the log line that
// explains it. A budget that expires before the platform gives up
// turns a diagnosable failure into a mystery.
const jobTimeout = 25 * time.Minute

// waitJob polls a job to completion.
//
// A poll that errors is not a job that failed: the per-second limit is
// signalled as an HTTP 500, and a container build outlasts any single
// transient. Errors are therefore retried until the deadline, and only
// a reported "Failed" state is treated as failure.
func waitJob(ctx context.Context, j *job.Client, jb models.Job) error {
	if jb.ID == 0 {
		return nil
	}
	deadline := time.Now().Add(jobTimeout)
	for {
		time.Sleep(throttle * 2)
		resp, err := j.Get(ctx, job.GetRequest{ID: jb.ID, Type: jb.Type})
		if err != nil {
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
//
// The family filter is not incidental. A container server reports both
// its IPv4 and its IPv6 address with primary set, so selecting on
// Primary alone returns whichever the API happened to list first — and
// the out-of-band checks need the v4 address specifically, because the
// host this journey runs on may have no v6 route at all.
func primaryIPv4(ctx context.Context, s *server.Client, name string) (string, error) {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return "", fmt.Errorf("Get(%s): %w", name, err)
	}
	for _, ip := range resp.Server.Ips {
		if ip.Primary && !strings.Contains(ip.IPAddr, ":") {
			return ip.IPAddr, nil
		}
	}
	return "", fmt.Errorf("Get(%s): no primary IPv4 reported", name)
}

// subject returns the container server the step should act on, and
// whether it is one this process created.
//
// A step run standalone needs a target, and SH_SERVER supplies it. The
// distinction matters for anything destructive: see deletable.
func subject(st *state) (string, error) {
	if st.name != "" {
		return st.name, nil
	}
	if v := os.Getenv("SH_SERVER"); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("no container server in state: run step 20 (provision) first, or set SH_SERVER")
}

// deletable returns only the servers this process provisioned.
//
// SH_SERVER is deliberately not consulted. Naming an existing server so
// that a read step has something to look at must never be the same act
// as marking it for deletion; in the server journey an earlier version
// conflated the two and attached a firewall to a real host.
func deletable(st *state) []string {
	return st.created
}

// ensureKey installs the key this run will use, generating one if the
// process has none.
//
// It is ephemeral by design: generated in memory, never written to
// disk, and useless once the SSH account is deleted at step 90.
func ensureKey(st *state) error {
	if st.publicKey != "" && len(st.privateKey) > 0 {
		journeyKey = st.privateKey
		return nil
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return fmt.Errorf("wrap public key: %w", err)
	}
	st.privateKey = priv
	st.publicKey = string(ssh.MarshalAuthorizedKey(sshPub))
	journeyKey = priv
	log.Printf("  generated an ephemeral ed25519 key for this run")
	return nil
}

// ensureAddress fills in the address when this step is run on its own.
//
// The out-of-band check needs somewhere to connect to, and a standalone
// run has no provision step to have recorded one.
func ensureAddress(ctx context.Context, c clients, st *state, name string) error {
	if st.ip != "" {
		return nil
	}
	time.Sleep(throttle)
	ip, err := primaryIPv4(ctx, c.server, name)
	if err != nil {
		return err
	}
	st.ip = ip
	log.Printf("  resolved %s to %s", name, st.ip)
	return nil
}

// generatedLabel builds a label whose derived server name is unique.
//
// Nine characters is the whole budget: the platform keeps only that
// many after the "ch-" prefix. "gosh" identifies the journey and the
// remaining five carry the entropy.
func generatedLabel() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		// A fixed label still works for a first run, and a collision
		// reports itself from the API rather than silently.
		return "gosh-cc"
	}
	return "gosh" + hex.EncodeToString(b)[:5]
}

// sshUserLabel returns a short label for naming account-level objects
// after this run.
//
// The SSH key is registered on the account rather than on the server,
// so it outlives the container unless something deletes it. Naming it
// after the run is what lets step 90 find and remove it.
func (st *state) sshUserLabel() string {
	if st.sshUser != "" {
		return st.sshUser
	}
	return "unnamed"
}

// randomSuffix returns a short hex string for naming per-run objects.
//
// Every name this journey creates is randomised, because the journey
// collides with itself otherwise: a database, a label or a username
// left behind by a failed run makes the next run fail on the name
// rather than on whatever it was testing.
func randomSuffix() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(b)
}
