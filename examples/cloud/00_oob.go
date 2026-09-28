package main

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Out-of-band verification helpers.
//
// # Some of these are copies, and that is deliberate
//
// Copied from examples/server, unchanged in substance: tcpReachable and
// waitReachability from 00_fwverify.go, signerFor from 00_ssh.go, and
// exists from 47_snapshot.go. sshRunAs is that journey's sshRun with
// the account passed in rather than derived, because a container SSH
// account is whatever was handed to cloud/ssh/user.Add and there is
// nothing to derive.
//
// New here, with no counterpart there: httpGetStatus, gatewayError,
// curlFromHost and waitCurlFromHost. A virtual server has no proxy in
// front of it and no Docker network behind it, so the server journey
// never needed them.
//
// Both journeys are package main in separate directories, so Go offers
// no way to share the copied ones short of extracting an importable
// package — which would mean editing the server journey from inside a
// cloud change. They are gathered in one file so the duplication is
// visible and that extraction is mechanical rather than archaeological.
// If you fix one of the copies here, fix it there too.
//
// The version of exists below is the one to copy from: an earlier form
// ran "test -e X && echo present || echo absent" and read the word from
// stdout, which reports "absent" identically whether the file is gone
// or the SSH session never opened. It checks the exit status instead.

// probeTimeout bounds one connection attempt.
const probeTimeout = 5 * time.Second

// probeSettle bounds how long a reachability or serving change is
// waited for.
//
// It is a variable rather than a constant so the tests covering the
// polling helpers can shorten it. Waiting the real ninety seconds to
// establish that a helper gives up made the unit suite slower than
// every other package combined, which is how a test stops being run.
var probeSettle = 90 * time.Second

// sshDialTimeout bounds a single SSH connection attempt.
const sshDialTimeout = 15 * time.Second

// tcpReachable reports whether a TCP connection to host:port completes.
//
// This is the whole point of the out-of-band checks: it opens a real
// socket from wherever the example runs, so it answers "is the port
// open" rather than "does the control plane believe the port is open".
func tcpReachable(host, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), probeTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// waitReachability polls until host:port matches want, or gives up.
// It returns whether the wanted state was reached and how long it took.
func waitReachability(host, port string, want bool) (bool, time.Duration) {
	start := time.Now()
	deadline := start.Add(probeSettle)
	for {
		if tcpReachable(host, port) == want {
			return true, time.Since(start)
		}
		if time.Now().After(deadline) {
			return false, time.Since(start)
		}
		time.Sleep(3 * time.Second)
	}
}

// journeyKey holds the private key this process generated, so the SSH
// helpers can be called from steps that do not thread state through.
var journeyKey ed25519.PrivateKey

// signerFor builds an ssh.Signer from the journey's key.
func signerFor() (ssh.Signer, error) {
	if len(journeyKey) == 0 {
		return nil, fmt.Errorf("no SSH key available: run step 30 (sshuser) in the same process")
	}
	return ssh.NewSignerFromKey(journeyKey)
}

// sshRunAs opens a session to addr as user and runs script, returning
// combined output.
//
// This differs from the server journey's sshRun in taking the account
// explicitly. A virtual server's login account is derived from its
// product family and distro; a Cloud Container's is whatever username
// was passed to cloud/ssh/user.Add, so there is nothing to derive.
//
// Host keys are deliberately not verified. The container server was
// created moments ago and is deleted at the end of the run, so there is
// no known_hosts entry to check against and nothing durable to protect.
// Do not copy this callback into anything that outlives a test: it
// accepts any host key, which means it cannot detect interception.
func sshRunAs(user, addr, script string) (string, error) {
	signer, err := signerFor()
	if err != nil {
		return "", err
	}

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // ephemeral throwaway host; see doc comment
		Timeout:         sshDialTimeout,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), cfg)
	if err != nil {
		return "", fmt.Errorf("dial %s as %s: %w", addr, user, err)
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session on %s: %w", addr, err)
	}
	defer func() { _ = session.Close() }()

	var out bytes.Buffer
	session.Stdout = &out
	session.Stderr = &out
	if err := session.Run(script); err != nil {
		return out.String(), fmt.Errorf("run on %s: %w", addr, err)
	}
	return out.String(), nil
}

// exists reports whether path exists on the host, by exit status.
//
// Read the file header before changing this: the exit status is the
// signal, precisely so that a failed connection cannot be mistaken for
// a missing file.
func exists(user, addr, path string) bool {
	_, err := sshRunAs(user, addr, "test -e "+path)
	return err == nil
}

// httpGetStatus performs a GET against url and returns the status code
// and a bounded prefix of the body.
//
// A container server publishes tcp/80, and nginx-proxy routes on the
// VIRTUAL_HOST environment variable a stack sets, so a stack deployed
// by this journey can be fetched from outside over the wildcard
// hostname. That is the strongest check available here: it proves the
// container is running and serving, which the stack's reported state
// does not.
func httpGetStatus(url string) (int, string, error) {
	c := &http.Client{Timeout: probeTimeout}
	resp, err := c.Get(url) //nolint:noctx // bounded by the client timeout
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(bytes.TrimSpace(body)), nil
}

// gatewayError reports a status that means the proxy never reached a
// container.
//
// This distinction is the whole value of the HTTP check. The host's
// reverse proxy answers on port 80 whether or not anything is behind
// it, and returns its own 503 page when nothing is routable — so
// "something answered" is satisfied by a stack that deployed and never
// started. A first version of this check accepted any status and
// passed on exactly that 503, while its own comment claimed to catch
// the case.
//
// A 403 or a 404 is not a failure here: those come from the container
// itself, which is what is being established. What it serves is step
// 50's business.
func gatewayError(code int) bool {
	return code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusGatewayTimeout
}

// curlFromHost fetches a URL from inside the container server, over
// SSH, and returns the status code and body.
//
// # Why this is the strongest check in the journey
//
// Fetching a stack's public hostname tests two things at once: whether
// the container serves, and whether the host's proxy routes to it. On a
// container server built by this journey the second is false — see the
// note in step 30 — so the public fetch fails for a reason that has
// nothing to do with the stack.
//
// Curling the container by its stack name from inside the host talks to
// it directly over the Docker network. A 200 here is the container
// itself answering, with no proxy in the path, which is exactly the
// thing a deploy needs to have established. It requires a shell, which
// is why the stack's serving proof lives in step 40 rather than step
// 30.
func curlFromHost(user, addr, url string) (int, string, error) {
	// -s so curl says nothing on its own, then the status code and the
	// body separated by a marker, so one command yields both.
	script := fmt.Sprintf(
		`curl -s -o /tmp/gosh-body -w '%%{http_code}' --max-time 8 %s; echo; cat /tmp/gosh-body; rm -f /tmp/gosh-body`,
		url)
	out, err := sshRunAs(user, addr, script)
	if err != nil {
		return 0, out, err
	}

	code, body, found := strings.Cut(strings.TrimLeft(out, "\n"), "\n")
	if !found {
		return 0, out, fmt.Errorf("could not read a status code out of %q", out)
	}
	n, convErr := strconv.Atoi(strings.TrimSpace(code))
	if convErr != nil {
		return 0, body, fmt.Errorf("curl reported %q, which is not a status code", code)
	}
	return n, body, nil
}

// waitCurlFromHost polls curlFromHost until the container answers.
func waitCurlFromHost(user, addr, url string) (int, string, time.Duration) {
	start := time.Now()
	deadline := start.Add(probeSettle)
	last := 0
	for {
		code, body, err := curlFromHost(user, addr, url)
		if err == nil && code != 0 && !gatewayError(code) {
			return code, body, time.Since(start)
		}
		if code != 0 {
			last = code
		}
		if time.Now().After(deadline) {
			return last, "", time.Since(start)
		}
		time.Sleep(3 * time.Second)
	}
}
