package gosh

import (
	"context"
	"fmt"
	"testing"
)

func TestNewClient(t *testing.T) {
	apikey := ""
	clientID := ""

	c := NewClient(apikey, clientID)

	ctx := context.Background()

	// CREATE Server
	resp, err := c.Servers.Create(ctx, &ServerCreateRequest{
		Label:       "GOSH",
		Location:    "SHQLIN",
		ProductCode: "XENLIT",
		Image:       "ubuntu-focal.amd64",
		Params: ParamsOptions{
			SSHKeys: []string{"ssh-rsa *****"},
		},
	})

	// GET Server
	// resp, err := c.Servers.Get(ctx, "gosh")

	// UPDATE Server
	// err := c.Servers.Update(ctx, &ServerUpdateRequest{Name: "gosh", Label: "ttgosh"})

	// UPGRADE Server
	// err := c.Servers.Upgrade(ctx, &ServerUpgradeRequest{Name: "gosh", Plan: "XENPRO"})

	// COMMIT CHANGES Server
	// resp, err := c.Servers.CommitChanges(ctx, "gosh")

	// DELETE Server
	// resp, err := c.Servers.Delete(ctx, "gosh")

	if err != nil {
		fmt.Println(err)
	}

	job, err := c.Jobs.Get(ctx, resp.Return.JobID)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(job)
}
