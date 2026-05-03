// Program mail-list-all is a read-only validation of mail.ListAll
// against a real mail server. Lists every domain on the server,
// then for the first domain calls ListAll and breaks the records
// out by Type (mailbox / alias / forward).
//
// Required env: SH_API_KEY, SH_CLIENT_ID, SH_MAIL_SERVER.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/mail"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("mail-list-all: %v", err)
	}
}

func run() error {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	mailServer := os.Getenv("SH_MAIL_SERVER")
	if apiKey == "" || clientID == "" || mailServer == "" {
		return fmt.Errorf("SH_API_KEY, SH_CLIENT_ID, SH_MAIL_SERVER required")
	}
	c, err := api.New(apiKey, clientID)
	if err != nil {
		return fmt.Errorf("api.New: %w", err)
	}

	mc := mail.New(c)
	ctx := context.Background()

	// 1. Discover the domains on the server.
	domains, err := mc.ListDomains(ctx, mail.ListDomainsOptions{
		ServerOptions: mail.ServerOptions{ServerName: mailServer},
	})
	if err != nil {
		return fmt.Errorf("ListDomains: %w", err)
	}
	log.Printf("✓ %s: %d domain(s)", mailServer, len(domains.Return))
	if len(domains.Return) == 0 {
		log.Printf("  no domains on this server — nothing to ListAll against")
		return nil
	}
	for _, d := range domains.Return {
		log.Printf("  %s  state=%s  accts=%d aliases=%d fwds=%d",
			d.Domain, d.State, d.Accounts, d.Nicknames, d.Forwarders)
	}

	// 2. ListAll against the first domain.
	target := domains.Return[0].Domain
	log.Printf("==> ListAll(%s, %s)", mailServer, target)
	all, err := mc.ListAll(ctx, mail.ListAllOptions{
		ServerOptions: mail.ServerOptions{ServerName: mailServer},
		Domain:        target,
	})
	if err != nil {
		return fmt.Errorf("ListAll: %w", err)
	}
	log.Printf("✓ %d record(s) total", len(all.Return))
	mailboxes, aliases, forwards := 0, 0, 0
	for _, r := range all.Return {
		switch r.Type {
		case 0:
			mailboxes++
			log.Printf("  [mailbox] %s  username=%s  label=%q",
				r.EmailAddr, r.Username, r.Label)
		case 1:
			aliases++
			log.Printf("  [alias  ] %s  -> %s", r.EmailAddr, r.Destination)
		case 2:
			forwards++
			log.Printf("  [forward] %s  -> %s", r.EmailAddr, r.Destination)
		default:
			log.Printf("  [type=%d ] %s  (unexpected type)", r.Type, r.EmailAddr)
		}
	}
	log.Printf("==> breakdown: %d mailboxes, %d aliases, %d forwards",
		mailboxes, aliases, forwards)
	return nil
}
