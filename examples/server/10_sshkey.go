package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"time"

	sshkey "github.com/sitehostnz/gosh/pkg/api/ssh/key"
	"golang.org/x/crypto/ssh"
)

// stepSSHKey generates an ephemeral keypair and registers the public
// half with the account.
//
// This has to happen before provisioning. A key cannot be injected into
// a server after the fact, so a server provisioned without one is
// reachable only by the password returned once in the provision
// response — which is no use to automation.
//
// Registration is not optional, and it is not the whole story either.
// Provisioning takes the public key *content* in params[ssh_keys][],
// not the id this call returns — but it validates that content against
// the account's registered keys, so both steps are needed. Skip the
// registration and the provision is rejected with:
//
//	One or more of the given SSH keys are not present in our system.
//	Please add them to your account, before attempting to use them
//	during a provision.
//
// Pass the id instead of the content and it fails the same way, since
// an id is not key material.
//
// The private half never touches disk. It lives in state for step 60
// and dies with the process.
func stepSSHKey(ctx context.Context, c clients, st *state) error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}
	authorized := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))

	label := fmt.Sprintf("gosh-journey-ephemeral-%d", time.Now().Unix())
	time.Sleep(throttle)
	resp, err := c.key.Create(ctx, sshkey.CreateRequest{
		Label:   label,
		Content: authorized,
	})
	if err != nil {
		return fmt.Errorf("ssh/key.Create: %w", err)
	}
	if resp.Return.KeyID == "" {
		return fmt.Errorf("ssh/key.Create: no key id returned")
	}

	st.keyID = resp.Return.KeyID
	st.publicKey = authorized
	st.privateKey = priv

	log.Printf("✓ registered ephemeral key %s (id %s)", label, st.keyID)
	log.Printf("  the private half stays in memory; step 90 deletes the registered key")
	return nil
}
