// Program redirect is a read-only validation of
// redirect.ListRedirects. Walks the nested map shape returned by
// /redirect/list_redirects.json and prints every rule grouped by
// domain.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/redirect"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("redirect: %v", err)
	}
}

func run() error {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return fmt.Errorf("SH_API_KEY and SH_CLIENT_ID required")
	}
	c, err := api.New(apiKey, clientID)
	if err != nil {
		return fmt.Errorf("api.New: %w", err)
	}

	resp, err := redirect.New(c).ListRedirects(context.Background(),
		redirect.ListRedirectsRequest{PageSize: 100})
	if err != nil {
		return fmt.Errorf("ListRedirects: %w", err)
	}

	total := 0
	for _, rules := range resp.Return {
		total += len(rules)
	}
	log.Printf("✓ %d domain(s), %d total redirect rule(s)", len(resp.Return), total)
	for domain, rules := range resp.Return {
		log.Printf("  %s", domain)
		for source, rule := range rules {
			log.Printf("    %d  %s -> %s", rule.Type, source, rule.Destination)
		}
	}
	return nil
}
