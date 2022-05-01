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
	//resp, err := c.Servers.Create(ctx, &ServerCreateRequest{
	//	Label:       "API10",
	//	Location:    "SHQLIN",
	//	ProductCode: "XENLIT",
	//	Image:       "ubuntu-focal.amd64",
	//})

	// GET Server
	//resp, err := c.Servers.Get(ctx, "ch-testing")

	// DELETE Server
	resp, err := c.Servers.Delete(ctx, "api10")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)
}
