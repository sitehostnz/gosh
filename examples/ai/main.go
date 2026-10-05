// Program ai walks the lifecycle of a SiteHost AI token — issue,
// inspect, reconfigure, revoke — and checks every step against the AI
// platform itself, not only against the API that made the change.
//
// # Why the checks go to ai.sitehost.nz
//
// Asking the control plane whether it applied a change proves a record
// was written, not that anything is enforced. So each configuration
// change is followed by a real request with the token:
//
//   - capabilities: an inference-only token must be refused on
//     /v1/embeddings, and accepted there once embedding is granted;
//   - allowlist: a token locked to an address that is not ours must be
//     refused, and accepted again once the allowlist is cleared;
//   - revocation: a token that was just answering must stop.
//
// Every refusal is bracketed by a success with the same token. Without
// that, a refusal proves nothing — a token that never worked would
// pass every "is it blocked?" check.
//
// # It writes
//
// This journey creates one token and deletes it. It refuses to run the
// write steps unless SH_EXAMPLE_ALLOW_PROVISION=1; without it, it lists
// the account's tokens and exits. The token it creates is revoked on
// failure and on interrupt (SIGINT, SIGTERM), by its recorded key id —
// never by matching a label. A SIGKILL cannot be caught: the label is
// logged before the token is issued so it can be found and removed.
// The AI platform calls it makes are billable usage: about a dozen,
// each a few tokens long.
//
// The token secret is never logged.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
// Optional: SH_BASE_URL, SH_RECORD_DIR, SH_EXAMPLE_ALLOW_PROVISION,
// SH_AI_BASE_URL (default https://ai.sitehost.nz/v1),
// SH_AI_CHAT_MODEL, SH_AI_EMBED_MODEL.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/sitehostnz/gosh/internal/recorder"
	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/ai/token"
)

// testNetIP is from 192.0.2.0/24 (RFC 5737, TEST-NET-1), which is never
// routed, so it cannot be the address this journey runs from.
const testNetIP = "192.0.2.1"

func main() {
	if err := run(); err != nil {
		log.Fatalf("ai: %v", err)
	}
}

func run() error {
	// An interrupt cancels the context rather than killing the process,
	// so the in-flight call fails, run returns, and the deferred
	// revocation below still runs.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tc, err := newTokenClient()
	if err != nil {
		return err
	}

	stepf("discover")
	list, err := tc.List(ctx)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	log.Printf("  ✓ %d token(s) on the account (total_items %d)", len(list.Return.Tokens), list.Return.TotalItems)

	if os.Getenv("SH_EXAMPLE_ALLOW_PROVISION") != "1" {
		log.Printf("  SKIP lifecycle: set SH_EXAMPLE_ALLOW_PROVISION=1 to issue and revoke a token")
		return nil
	}

	ai := newAIClient()

	stepf("add (inference only)")
	label := "gosh-example-ai-" + strconv.FormatInt(time.Now().Unix(), 10)
	// Announced before the call: if Add reaches the API and the response
	// is lost, this label is the only handle on the token it made.
	log.Printf("  issuing token labelled %s", label)
	added, err := tc.Add(ctx, token.AddRequest{
		Label:        label,
		Capabilities: []string{token.CapabilityInference},
	})
	if err != nil {
		return fmt.Errorf("add: %w", err)
	}
	keyID := added.Return.KeyID
	ai.token = added.Return.Token
	if keyID == "" || ai.token == "" {
		return fmt.Errorf("add: succeeded without a key_id or token (key_id %q, token length %d)", keyID, len(ai.token))
	}
	log.Printf("  ✓ key_id=%s, token of %d chars", keyID, len(ai.token))

	// From here on the token exists, so every return revokes it unless
	// the journey's own delete step already has.
	revoked := false
	defer func() {
		if revoked {
			return
		}
		if _, err := tc.Delete(context.Background(), token.DeleteRequest{KeyID: keyID}); err != nil {
			log.Printf("  ✗ cleanup: could not revoke %s, remove it by hand: %v", keyID, err)
			return
		}
		log.Printf("  cleanup: revoked %s", keyID)
	}()

	if err := lifecycle(ctx, tc, ai, keyID, label, &revoked); err != nil {
		return err
	}
	log.Printf("✓ all checks passed")
	return nil
}

// step is one named stage of the journey and the checks it runs.
type step struct {
	name string
	run  []func() error
}

// lifecycle runs every check that needs the issued token, in order.
// Its final step revokes the token.
func lifecycle(ctx context.Context, tc *token.Client, ai aiClient, keyID, label string, revoked *bool) error {
	for _, s := range steps(ctx, tc, ai, keyID, label, revoked) {
		stepf("%s", s.name)
		for _, check := range s.run {
			if err := check(); err != nil {
				return err
			}
		}
	}
	return nil
}

// steps lists the journey. Each refusal it expects is preceded, within
// the same token's history, by a success on the same call.
func steps(ctx context.Context, tc *token.Client, ai aiClient, keyID, label string, revoked *bool) []step {
	both := []string{token.CapabilityInference, token.CapabilityEmbedding}
	renamed := label + "-renamed"

	update := func(what string, req token.UpdateRequest) error {
		req.KeyID = keyID
		if _, err := tc.Update(ctx, req); err != nil {
			return fmt.Errorf("update %s: %w", what, err)
		}
		return nil
	}

	return []step{
		{"get", []func() error{
			func() error { return expectConfig(ctx, tc, keyID, label, []string{token.CapabilityInference}, nil) },
		}},
		{"list", []func() error{
			func() error { return expectListed(ctx, tc, keyID, true) },
		}},
		{"models are served (precondition for every check below)", []func() error{
			func() error { return ai.expectModels(ctx) },
		}},
		{"capabilities are enforced by the AI platform", []func() error{
			func() error { return ai.expect(ctx, "chat", ai.chat, http.StatusOK) },
			func() error {
				return ai.expect(ctx, "embeddings without the capability", ai.embed, http.StatusForbidden)
			},
		}},
		{"update: grant embedding", []func() error{
			func() error { return update("capabilities", token.UpdateRequest{Capabilities: both}) },
			func() error { return expectConfig(ctx, tc, keyID, label, both, nil) },
			func() error { return ai.expect(ctx, "embeddings with the capability", ai.embed, http.StatusOK) },
		}},
		{"update: lock to " + testNetIP + " (not us)", []func() error{
			func() error { return update("allowed_ips", token.UpdateRequest{AllowedIPs: []string{testNetIP}}) },
			func() error { return expectConfig(ctx, tc, keyID, label, both, []string{testNetIP}) },
			func() error {
				return ai.expect(ctx, "chat from outside the allowlist", ai.chat, http.StatusForbidden)
			},
		}},
		{"update: clear the allowlist", []func() error{
			func() error { return update("clear allowed_ips", token.UpdateRequest{AllowedIPs: []string{}}) },
			func() error { return expectConfig(ctx, tc, keyID, label, both, nil) },
			func() error { return ai.expect(ctx, "chat with the allowlist cleared", ai.chat, http.StatusOK) },
		}},
		{"update: label only", []func() error{
			func() error { return update("label", token.UpdateRequest{Label: renamed}) },
			func() error { return expectConfig(ctx, tc, keyID, renamed, both, nil) },
			// Also the success that brackets the revocation check: the
			// token must be answering immediately before it is revoked.
			func() error { return ai.expect(ctx, "chat after a label-only change", ai.chat, http.StatusOK) },
		}},
		{"delete", []func() error{
			func() error { return revoke(ctx, tc, keyID, revoked) },
			func() error { return ai.expect(ctx, "chat after revocation", ai.chat, http.StatusUnauthorized) },
			func() error { return expectGone(ctx, tc, keyID) },
			func() error { return expectListed(ctx, tc, keyID, false) },
		}},
	}
}

// revoke deletes the token and checks the API says so. It sets
// *revoked once the delete is accepted, so that a check failing after
// it does not send cleanup after a token that is already gone.
func revoke(ctx context.Context, tc *token.Client, keyID string, revoked *bool) error {
	del, err := tc.Delete(ctx, token.DeleteRequest{KeyID: keyID})
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	*revoked = true
	if !del.Return.Deleted {
		return errors.New("delete: accepted but reported deleted=false")
	}
	return nil
}

// expectGone checks the control plane no longer returns the token.
func expectGone(ctx context.Context, tc *token.Client, keyID string) error {
	if _, err := tc.Get(ctx, token.GetRequest{KeyID: keyID}); err == nil {
		return errors.New("get: a revoked token is still returned")
	}
	log.Printf("  ✓ get refuses the revoked token")
	return nil
}

// expectConfig reads the token back and checks its label, capabilities
// and allowlist. A nil allowedIPs means "no restriction".
func expectConfig(ctx context.Context, tc *token.Client, keyID, label string, caps, allowedIPs []string) error {
	got, err := tc.Get(ctx, token.GetRequest{KeyID: keyID})
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	t := got.Return
	if t.KeyID != keyID {
		return fmt.Errorf("get: key_id %q, want %q", t.KeyID, keyID)
	}
	if t.Label != label {
		return fmt.Errorf("get: label %q, want %q", t.Label, label)
	}
	if !sameSet(t.Config.Capabilities, caps) {
		return fmt.Errorf("get: capabilities %v, want %v", t.Config.Capabilities, caps)
	}
	if !sameSet(t.Config.AllowedIPs, allowedIPs) {
		return fmt.Errorf("get: allowed_ips %v, want %v", t.Config.AllowedIPs, allowedIPs)
	}
	log.Printf("  ✓ control plane: capabilities %v, %d allowed IP(s)", t.Config.Capabilities, len(t.Config.AllowedIPs))
	return nil
}

// expectListed checks whether keyID appears in the account's tokens.
func expectListed(ctx context.Context, tc *token.Client, keyID string, want bool) error {
	list, err := tc.List(ctx)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	found := slices.ContainsFunc(list.Return.Tokens, func(s token.Summary) bool { return s.KeyID == keyID })
	if found != want {
		return fmt.Errorf("list: key %s listed=%v, want %v (%d tokens)", keyID, found, want, len(list.Return.Tokens))
	}
	log.Printf("  ✓ listed=%v among %d token(s)", found, len(list.Return.Tokens))
	return nil
}

// aiClient calls the AI platform with the token under test.
type aiClient struct {
	base       string
	token      string
	chatModel  string
	embedModel string
	http       *http.Client
}

// newAIClient configures the AI platform client from the environment.
// The token is filled in once one has been issued.
func newAIClient() aiClient {
	return aiClient{
		base:       envOr("SH_AI_BASE_URL", "https://ai.sitehost.nz/v1"),
		chatModel:  envOr("SH_AI_CHAT_MODEL", "Qwen/Qwen3.6-27B"),
		embedModel: envOr("SH_AI_EMBED_MODEL", "Qwen/Qwen3-VL-Embedding-8B"),
		http:       &http.Client{Timeout: 60 * time.Second},
	}
}

// models lists the served model ids.
func (a aiClient) models(ctx context.Context) ([]string, error) {
	status, body, err := a.do(ctx, http.MethodGet, "/models", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("GET /models: HTTP %d", status)
	}
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("GET /models: %w", err)
	}
	ids := make([]string, 0, len(out.Data))
	for _, m := range out.Data {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

// expectModels checks both configured models are served, so that a
// refusal later means what it says rather than "no such model".
func (a aiClient) expectModels(ctx context.Context) error {
	models, err := a.models(ctx)
	if err != nil {
		return err
	}
	for _, m := range []string{a.chatModel, a.embedModel} {
		if !slices.Contains(models, m) {
			return fmt.Errorf("model %q is not served (%d models listed); set SH_AI_CHAT_MODEL / SH_AI_EMBED_MODEL", m, len(models))
		}
	}
	log.Printf("  ✓ %s and %s are served", a.chatModel, a.embedModel)
	return nil
}

// chat makes the smallest possible chat completion.
func (a aiClient) chat(ctx context.Context) (int, error) {
	status, _, err := a.do(ctx, http.MethodPost, "/chat/completions", map[string]any{
		"model":      a.chatModel,
		"messages":   []map[string]string{{"role": "user", "content": "Say ok"}},
		"max_tokens": 4,
		// Qwen thinks by default, which would spend the token budget
		// before answering; the check only needs a status code.
		"chat_template_kwargs": map[string]bool{"enable_thinking": false},
	})
	return status, err
}

// embed makes the smallest possible embeddings request.
func (a aiClient) embed(ctx context.Context) (int, error) {
	status, _, err := a.do(ctx, http.MethodPost, "/embeddings", map[string]any{
		"model": a.embedModel,
		"input": "ok",
	})
	return status, err
}

// expect calls fn and checks the HTTP status it got. The status is the
// check, read directly: a body is not parsed for words.
func (a aiClient) expect(ctx context.Context, what string, fn func(context.Context) (int, error), want int) error {
	got, err := fn(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if got != want {
		return fmt.Errorf("%s: HTTP %d, want %d", what, got, want)
	}
	log.Printf("  ✓ ai platform: %s → HTTP %d", what, got)
	return nil
}

func (a aiClient) do(ctx context.Context, method, path string, payload any) (int, []byte, error) {
	var body *bytes.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(b)
	} else {
		body = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, a.base+path, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, buf.Bytes(), nil
}

// newTokenClient builds the token client, optionally recording every
// call to the SiteHost API (not to the AI platform).
func newTokenClient() (*token.Client, error) {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return nil, errors.New("SH_API_KEY and SH_CLIENT_ID required")
	}

	var opts []api.ClientOpt
	if base := os.Getenv("SH_BASE_URL"); base != "" {
		opts = append(opts, api.SetBaseURL(base))
	}
	if dir := os.Getenv("SH_RECORD_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("SH_RECORD_DIR: %w", err)
		}
		log.Printf("  recording every call to %s", dir)
		opts = append(opts, api.SetTransport(recorder.New(dir, nil)))
	}

	c, err := api.New(apiKey, clientID, opts...)
	if err != nil {
		return nil, fmt.Errorf("api.New: %w", err)
	}
	return token.New(c), nil
}

// sameSet reports whether a and b hold the same values, ignoring order.
// nil and empty are equal.
func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func stepf(format string, args ...any) {
	log.Printf("── "+format, args...)
}
