# Gosh

Gosh is a Go client library for accessing the [SiteHost v1.3 API](https://docs.sitehost.nz/api/v1.3/).

## Installation

```sh
go get -u https://github.com/sitehostnz/gosh
```

## Example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/server"
)

func main() {
	apiKey := os.Getenv("SH_API_KEY")
	clientId := os.Getenv("SH_CLIENT_ID")

	client := api.NewClient(apiKey, clientId)
	ctx := context.Background()

	instance := server.New(client)

	opts := server.CreateRequest{
		Label:       "goshserver",
		Location:    "AKLCITY",
		ProductCode: "XENLIT",
		Image:       "ubuntu-jammy-pvh.amd64",
		Params: server.ParamsOptions{
			SSHKeys: []string{"ssh-rsa ..."},
		},
	}

	server, err := instance.Create(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", server)
}
```

## Documentation

The structure of this library closely mirrors that of our API, so the
[API documentation](https://docs.sitehost.nz/api/v1.3/) should be your first
point of reference.

## Contributing

If you're interested in contributing to our project:
- Start by reading our [style guide](https://github.com/sitehostnz/go-style-guide/blob/master/style.md) and [contributing guide](/docs/CONTRIBUTING.md).
- Explore our [issues](https://github.com/sitehostnz/gosh/issues).
- Or send us feature PRs.

## Licence

Gosh is distributed under the terms of the [MIT](./LICENSE.md) licence.
