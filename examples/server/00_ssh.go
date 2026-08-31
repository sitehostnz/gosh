package main

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
	"golang.org/x/crypto/ssh"
)

// sshDialTimeout bounds a single connection attempt.
const sshDialTimeout = 15 * time.Second

// journeyLoginUser is the account resolved from the provisioned
// server's product family and distro, set by step 30. Package-level
// because the ssh helpers are called from steps that do not thread
// state through.
var journeyLoginUser string

// sshUser is the account to log in as.
//
// The API does not report this, and it depends on both the product
// family and the distro — the same Ubuntu image logs in as "ubuntu" on
// high-performance and "root" on legacy Xen (LINVPS). The
// standard-performance (SVS) tier was not tested, and
// server.LoginUserFor deliberately returns not-ok for it rather than
// guessing. Step 30 resolves
// it through server.LoginUserFor from what the API reports about the
// server it just created. SH_SSH_USER overrides.
func sshUser() string {
	if v := os.Getenv("SH_SSH_USER"); v != "" {
		return v
	}
	if journeyLoginUser != "" {
		return journeyLoginUser
	}
	// No provision in this process to learn from. Say what is being
	// assumed rather than failing obscurely later.
	log.Printf("  no login account resolved; assuming %q — set SH_SSH_USER if that is wrong", server.LoginUserRoot)
	return server.LoginUserRoot
}

// sshRun opens a session to addr using the journey's ephemeral key and
// runs script, returning combined output.
//
// Host keys are deliberately not verified. These servers were created
// moments ago and are deleted at the end of the run, so there is no
// known_hosts entry to check against and nothing durable to protect.
// Do not copy this callback into anything that outlives a test: it
// accepts any host key, which means it cannot detect interception.
func sshRun(addr, script string) (string, error) {
	signer, err := signerFor()
	if err != nil {
		return "", err
	}

	cfg := &ssh.ClientConfig{
		User:            sshUser(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // ephemeral throwaway hosts; see doc comment
		Timeout:         sshDialTimeout,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), cfg)
	if err != nil {
		return "", fmt.Errorf("dial %s as %s: %w", addr, cfg.User, err)
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

// journeyKey holds the private key for the current process, set by step
// 10. It is package-level because the ssh helpers are called from steps
// that do not thread state through.
var journeyKey ed25519.PrivateKey

// requireKey checks a key is reachable and installs the in-process one
// when there is one.
//
// Either source will do: the key this process generated at step 10, or
// one the operator points at with SH_SSH_KEY_FILE. Rejecting on the
// first alone made the second unreachable — the error told you to set
// a variable that was never consulted, so following its own advice
// produced the identical error. Steps 50 and 80 are the two that need
// SSH, so that made them the two that could not run standalone at all.
func requireKey(st *state, what string) error {
	if len(st.privateKey) == 0 && os.Getenv("SH_SSH_KEY_FILE") == "" {
		return fmt.Errorf("no SSH key available: %s needs the key from step 10 in the same process, or SH_SSH_KEY_FILE pointing at a key the servers already trust", what)
	}
	// Only overwrite the package-level key when this process has one;
	// assigning an empty key here would mask the file fallback.
	if len(st.privateKey) > 0 {
		journeyKey = st.privateKey
	}
	return nil
}

// signerFor builds an ssh.Signer from the journey's key.
func signerFor() (ssh.Signer, error) {
	if len(journeyKey) == 0 {
		if p := os.Getenv("SH_SSH_KEY_FILE"); p != "" {
			raw, err := os.ReadFile(p) //nolint:gosec // path is supplied by the operator running the example
			if err != nil {
				return nil, fmt.Errorf("read SH_SSH_KEY_FILE: %w", err)
			}
			return ssh.ParsePrivateKey(raw)
		}
		return nil, fmt.Errorf("no SSH key available: run step 10 in the same process, or set SH_SSH_KEY_FILE")
	}
	return ssh.NewSignerFromKey(journeyKey)
}
