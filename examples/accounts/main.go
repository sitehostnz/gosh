// Program accounts is a read-only validation of accounts.client.ListSubAccounts.
// Lists every sub-account under the calling client_id and prints
// id / name / balance / type. Useful as a discovery step for resellers
// driving the SDK against their full customer set.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sitehostnz/gosh/pkg/api"
	accountsClient "github.com/sitehostnz/gosh/pkg/api/accounts/client"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("accounts: %v", err)
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

	ctx := context.Background()
	resp, err := accountsClient.New(c).ListSubAccounts(ctx, accountsClient.ListSubAccountsRequest{
		PageSize: 100,
	})
	if err != nil {
		return fmt.Errorf("ListSubAccounts: %w", err)
	}
	log.Printf("✓ %d sub-account(s) returned (page 1, total %d, total_pages %d)",
		len(resp.Return.Accounts), resp.Return.TotalItems, resp.Return.TotalPages)
	for _, a := range resp.Return.Accounts {
		closed := a.Closed
		if closed == "0000-00-00" {
			closed = "(open)"
		}
		log.Printf("  id=%s name=%-30s balance=%-8s joined=%s closed=%s type=%s",
			a.ClientID, a.Name, a.AccountBalance, a.Joined, closed, a.AccountType)
	}
	return nil
}
