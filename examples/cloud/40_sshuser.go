package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	sshkey "github.com/sitehostnz/gosh/pkg/api/ssh/key"
)

// stepSSHUser creates a container SSH account and logs into it.
//
// # This is the journey's first real out-of-band proof
//
// cloud/ssh/user/add.json returns a job. The job completing means the
// platform believes it wrote the account. Asking the API afterwards
// whether the account exists confirms only that the same system still
// holds the same record — it cannot distinguish an account that works
// from a row in a table.
//
// So this step opens an SSH connection to the container server's
// address, as the account it just created, authenticating with the key
// pair generated at step 20, and runs a command. If the account was
// never really usable, the dial or the auth fails and the step fails
// with it. Nothing about that check goes through the control plane.
//
// # params[ssh_keys][] means something different here
//
// On server.Create the same parameter name carries public key
// *content*. On cloud/ssh/user it carries the *id* of a key already
// registered on the account through ssh/key/add.json. Nothing says so,
// and the two endpoints disagree about how badly they take it:
//
//   - cloud/ssh/user/update.json rejects content with "One or more of
//     the given SSH keys could not be found or you do not have access
//     to them."
//   - cloud/ssh/user/add.json returns a bare HTTP 500 with an empty
//     body — no JSON, no message. Removing the parameter makes the same
//     call succeed, which is how this was isolated.
//
// So this step registers the public key first and passes the id it gets
// back. That is also the only thing in this repository that exercises
// pkg/api/ssh/key, which had no coverage at all.
//
// # The request struct's tags do not describe what is sent
//
// AddRequest tags Username as "username" and Password as "password",
// but Client.Add builds params[password], params[containers][] and
// params[ssh_keys][] by hand and never encodes the struct. The tags are
// therefore inert: correct for the shape of the call, wrong about the
// parameter names. Anything that encoded this struct directly would
// send names the API ignores — which is the failure mode that has
// already produced four silent bugs in this SDK, so it is worth
// knowing the tags here are not the contract.
//
// # This step generates the key, and can therefore run standalone
//
// The key pair belongs here rather than to the provision step: on a
// Cloud Container it is registered against the SSH account, not the
// server, so provisioning has nothing to do with it. Generating it here
// is also what lets this step and the two after it run against a server
// named by SH_SERVER — the server journey learned that the hard way,
// where a check for a key from an earlier step made the standalone path
// unreachable and told you to set a variable it never read.
//
// # The container grant is required, not optional
//
// cloud/ssh/user/add.json rejects an account with neither containers
// nor volumes — "Please specify a valid list of volumes or
// containers." — so the website container from step 30 is granted at
// creation. That is also what makes the account useful: the grant is
// what gives it access to the container's files, which step 50 needs
// in order to write a page into the docroot.
func stepSSHUser(ctx context.Context, c clients, st *state) error {
	name, err := subject(st)
	if err != nil {
		return err
	}
	if st.stackName == "" {
		return fmt.Errorf("no container to grant: run step 30 (stack) first — ssh/user.Add rejects an account with no containers or volumes")
	}
	if err := ensureAddress(ctx, c, st, name); err != nil {
		return err
	}

	if err := createSSHUser(ctx, c, st, name); err != nil {
		return err
	}
	if err := confirmSSHUserReadable(ctx, c, name, st); err != nil {
		return err
	}
	if err := proveLogin(st); err != nil {
		return err
	}
	return proveContainerServes(st)
}

// createSSHUser registers the key and creates the account.
//
// The username is chosen before the key is registered so the key's
// label names this run, which is what step 90 matches on.
func createSSHUser(ctx context.Context, c clients, st *state, name string) error {
	st.sshUser = generatedUser()
	if err := ensureKey(st); err != nil {
		return err
	}
	if err := registerKey(ctx, c, st); err != nil {
		return err
	}

	password, err := generatedSecret()
	if err != nil {
		return err
	}

	time.Sleep(throttle)
	resp, err := c.sshers.Add(ctx, sshuser.AddRequest{
		ServerName: name,
		Username:   st.sshUser,
		Password:   password,
		SSHKeys:    []string{st.sshKeyID},
		Containers: []string{st.stackName},
	})
	if err != nil {
		return fmt.Errorf("ssh/user.Add(%s): %w", st.sshUser, err)
	}
	log.Printf("✓ requested ssh user %q on %s (job %d)", st.sshUser, name, resp.Return.ID)

	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("ssh/user.Add(%s): %w", st.sshUser, err)
	}
	return nil
}

// confirmSSHUserReadable checks the account is reported back.
//
// This is the control-plane half, and it is kept because a mismatch
// between what was asked for and what is reported is worth catching —
// but on its own it proves only that a record exists. proveLogin is the
// half that can fail for the right reasons.
func confirmSSHUserReadable(ctx context.Context, c clients, name string, st *state) error {
	time.Sleep(throttle)
	got, err := c.sshers.Get(ctx, sshuser.GetRequest{ServerName: name, Username: st.sshUser})
	if err != nil {
		return fmt.Errorf("ssh/user.Get(%s): %w", st.sshUser, err)
	}
	if got.Return.Username != st.sshUser {
		return fmt.Errorf("ssh/user.Get: username is %q, want %q", got.Return.Username, st.sshUser)
	}
	// Assert the grant landed rather than only counting it: an account
	// created with the wrong container would report a plausible 1 here.
	for _, ct := range got.Return.Containers {
		if ct == st.stackName {
			log.Printf("  control plane reports the account with the container granted: username=%q", got.Return.Username)
			return nil
		}
	}
	return fmt.Errorf("ssh/user.Get: container %q is not among the account's grants %v", st.stackName, got.Return.Containers)
}

// proveLogin logs in over SSH and runs a command.
//
// The assertion is that the shell reports the account this step just
// created. A session that opened as somebody else, or as nobody, would
// otherwise read as a pass.
//
// Everything short of that is retried rather than raised, because a
// newly created account goes through a window where the login already
// works and the identity does not resolve yet. Two forms of it were
// observed within minutes of each other: the command returning no
// output at all, and it returning a run of NUL bytes in place of the
// username. Both were transient, and both had been written up as
// failures — one of them with the NULs printed into the error, which is
// how it was recognised as a readiness problem rather than a wrong
// account.
//
// A genuinely wrong account never matches, so it still fails, with the
// last thing the shell said.
func proveLogin(st *state) error {
	deadline := time.Now().Add(probeSettle)
	var last string
	for {
		out, err := sshRunAs(st.sshUser, st.ip, "id -un && pwd")
		if err != nil {
			last = err.Error()
		} else {
			fields := strings.Fields(strings.ReplaceAll(out, "\x00", ""))
			if len(fields) > 0 && fields[0] == st.sshUser {
				log.Printf("✓ out of band: SSH login as %q succeeded, id=%s home=%s",
					st.sshUser, fields[0], strings.Join(fields[1:], " "))
				return nil
			}
			last = fmt.Sprintf("the shell reported %q", strings.Join(fields, " "))
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("SSH as %q never reported that account within %s: %s",
				st.sshUser, probeSettle, last)
		}
		time.Sleep(5 * time.Second)
	}
}

// generatedUser builds a username unlikely to collide.
//
// Container SSH usernames are namespaced per server, but a journey rerun
// against the same server would collide with itself, so the name is
// randomised rather than fixed.
func generatedUser() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Falling back to a fixed name is better than failing the step
		// here; a collision reports itself clearly from the API.
		return "goshjourney"
	}
	return "gosh" + hex.EncodeToString(b)
}

// generatedSecret builds a random password.
//
// Key authentication is what this journey uses, but the endpoint takes
// a password and sends params[password] whether or not one was set, so
// a random one is supplied rather than an empty string. It is never
// logged and never reused.
func generatedSecret() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// registerKey adds the journey's public key to the account and records
// the id.
//
// The id, not the key material, is what cloud/ssh/user.Add accepts.
// Passing the material there returns a bare HTTP 500, so this call is
// not optional plumbing — it is the difference between the step working
// and the step crashing the endpoint.
func registerKey(ctx context.Context, c clients, st *state) error {
	if st.sshKeyID != "" {
		return nil
	}

	time.Sleep(throttle)
	created, err := c.sshkeys.Create(ctx, sshkey.CreateRequest{
		Label:   "gosh-journey-" + st.sshUserLabel(),
		Content: st.publicKey,
	})
	if err != nil {
		return fmt.Errorf("ssh/key.Create: %w", err)
	}
	if created.Return.KeyID == "" {
		return fmt.Errorf("ssh/key.Create: returned no key id")
	}
	st.sshKeyID = created.Return.KeyID

	// Read it back and check the content matches what was sent. The id
	// is the only thing later calls use, so an id belonging to somebody
	// else's key would authenticate as a different key and fail the
	// login for a reason that looks like the account being broken.
	time.Sleep(throttle)
	got, err := c.sshkeys.Get(ctx, sshkey.GetRequest{ID: st.sshKeyID})
	if err != nil {
		return fmt.Errorf("ssh/key.Get(%s): %w", st.sshKeyID, err)
	}
	if strings.TrimSpace(got.Return.Content) != strings.TrimSpace(st.publicKey) {
		return fmt.Errorf("ssh/key.Get(%s): the registered key is not the one this run generated", st.sshKeyID)
	}
	log.Printf("✓ registered the journey key as id %s", st.sshKeyID)
	return nil
}

// proveContainerServes writes a marker into the container's docroot and
// fetches it from inside the host.
//
// This is where the stack deployed by step 30 is actually proven to
// work, and it is here rather than there because it needs a shell. One
// request establishes three things at once:
//
//   - the SSH account can write to the container's files, which is what
//     the container grant is for;
//   - the container is serving, not merely reported Up;
//   - it is serving the content this journey put there, so a default
//     page from some other container cannot be mistaken for a pass.
//
// The fetch goes to the stack name over the Docker network, not to the
// public hostname. See step 30 for why the public one cannot be used.
func proveContainerServes(st *state) error {
	docroot, err := findDocroot(st)
	if err != nil {
		return err
	}

	marker := "gosh-" + st.stackName + "-alive"
	target := docroot + "/gosh-probe.txt"

	// printf rather than a heredoc. A heredoc here created the file and
	// then exited 1, leaving a zero-byte probe and an error about the
	// write rather than about the shell — the payload is fine, the
	// delimiter handling over a non-interactive exec is not. One
	// command with a quoted argument has no delimiter to get wrong.
	script := fmt.Sprintf("printf %%s %s > %s", shellQuote(marker), shellQuote(target))
	if _, err := sshRunAs(st.sshUser, st.ip, script); err != nil {
		return fmt.Errorf("write a probe file to %s: %w", target, err)
	}
	if !exists(st.sshUser, st.ip, target) {
		return fmt.Errorf("wrote %s but it does not exist afterwards", target)
	}

	url := "http://" + st.stackName + "/gosh-probe.txt"
	code, body, took := waitCurlFromHost(st.sshUser, st.ip, url)
	switch {
	case code == 0:
		return fmt.Errorf("%s never answered from inside the host within %s — the container is not serving", url, probeSettle)
	case code != 200:
		return fmt.Errorf("%s answered HTTP %d from inside the host, want 200", url, code)
	case !strings.Contains(body, marker):
		return fmt.Errorf("%s answered 200 but did not return the marker this run wrote; got %q", url, firstLine(body))
	}
	log.Printf("✓ out of band: the container served its own marker over the Docker network after %s", took.Round(time.Second))
	return nil
}

// docrootCandidates are the paths a website container may serve from.
//
// The docroot is not reported by the API, so it is probed rather than
// assumed. On the image this journey deploys it is the first entry; the
// others are kept because the check costs one command and a wrong guess
// otherwise fails with a confusing write error.
var docrootCandidates = []string{
	"/container/application/public",
	"/container/application",
	"/var/www/html",
}

// findDocroot probes the candidate paths over SSH.
//
// A failure here is reported as such. Falling back to a control-plane
// check would turn a step that can fail into one that cannot, which is
// the failure mode this journey exists to avoid.
func findDocroot(st *state) (string, error) {
	for _, dir := range docrootCandidates {
		if exists(st.sshUser, st.ip, dir) {
			return dir, nil
		}
	}
	return "", fmt.Errorf("no docroot found on %s as %q among %v — cannot verify the container out of band",
		st.ip, st.sshUser, docrootCandidates)
}

// firstLine returns a short, single-line prefix of a body, for logging.
func firstLine(body string) string {
	line := body
	if i := strings.IndexAny(line, "\r\n"); i >= 0 {
		line = line[:i]
	}
	if len(line) > 96 {
		line = line[:96] + "…"
	}
	return strings.TrimSpace(line)
}
