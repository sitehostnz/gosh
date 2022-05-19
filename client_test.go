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
			SSHKeys: []string{"ssh-rsa *************"},
		},
	})

	// GET Server
	//resp, err := c.Servers.Get(ctx, "ch-testing")

	// UPDATE Server
	//err := c.Servers.Update(ctx, &ServerUpdateRequest{Name: "trtest", Label: "ttgosh"})
	//if err != nil {
	//	fmt.Println("ERROR", err)
	//} else {
	//	fmt.Println("OK")
	//}

	// DELETE Server
	//resp, err := c.Servers.Delete(ctx, "api10")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)

	job, err := c.Jobs.Get(ctx, resp.Return.JobID)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(job)

	//err := c.Servers.Upgrade(ctx, &ServerUpgradeRequest{Name: "gosh", Plan: "XENPRO"})
	//if err != nil {
	//fmt.Println(err)
	//} else {
	//	fmt.Println("Changed to PRO")
	//}

	//resp, err := c.Servers.CommitChanges(ctx, "gosh")
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Println(resp)
	//}
}
