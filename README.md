# gosh

Gosh is a Go client library for accessing the [SiteHost V1.1 API](https://docs.sitehost.nz/api/v1/)

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

	"github.com/sitehostnz/gosh"

	"log"
	"net/http"
	"os"
)

func main() {
  apiKey := os.Getenv("SH_APIKEY")
  clientId := os.Getenv("SH_CLIENTID")

  sitehostClient := gosh.NewClient(nil)

  sitehostClient.SetToken(apiKey, clientId)
  sitehostClient.SetDebug(true)

  server, err := sitehostClient.GetServer(context.Background(), "ch-server1")
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%v", server)
}
```

## Documentation

See our documentation for [detailed information about API v1.1](https://docs.sitehost.nz/api/v1/).