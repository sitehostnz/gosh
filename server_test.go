package gosh

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServer(t *testing.T) {

	apiKey := os.Getenv("SH_APIKEY")
	clientId := os.Getenv("SH_CLIENTID")
	serverName := "ch-gonzalo"

	c := NewClient(nil)
	c.SetDebug(true)
	c.SetToken(apiKey, clientId)

	ctx := context.Background()

	server, err := c.GetServer(ctx, serverName)

	assert.Nil(t, err, "expecting nil error")
	assert.NotNil(t, server, "expecting non-nil result")

	if err != nil {
		assert.Equal(t, serverName, server.Name, fmt.Sprintf("expecting server name: %s", serverName))
	}
}

func TestDeleteServer(t *testing.T) {

	apiKey := os.Getenv("SH_APIKEY")
	clientId := os.Getenv("SH_CLIENTID")
	serverName := "ch-gonzalo"

	c := NewClient(nil)
	c.SetDebug(true)
	c.SetToken(apiKey, clientId)

	ctx := context.Background()

	err := c.DeleteServer(ctx, serverName)

	assert.Nil(t, err, "expecting nil error")
}
