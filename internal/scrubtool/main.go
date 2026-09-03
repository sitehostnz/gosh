// Command scrubtool converts recordings into fixtures and reports what
// the SDK's types do not decode. Development tool; not part of the API.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sitehostnz/gosh/internal/recorder"
	"github.com/sitehostnz/gosh/internal/shapecheck"
	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	cloudserver "github.com/sitehostnz/gosh/pkg/api/cloud/server"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
	stackimage "github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
)

// types maps an endpoint to the response type that decodes it.
var types = map[string]func() any{
	"1.5/cloud/db/list_all.json":          func() any { return db.ListResponse{} },
	"1.5/cloud/db/get.json":               func() any { return db.GetResponse{} },
	"1.5/cloud/db/user/list_all.json":     func() any { return dbuser.ListResponse{} },
	"1.5/cloud/ssh/user/list_all.json":    func() any { return sshuser.ListResponse{} },
	"1.5/cloud/stack/list_all.json":       func() any { return stack.ListResponse{} },
	"1.5/cloud/stack/get.json":            func() any { return stack.GetResponse{} },
	"1.5/cloud/stack/image/list_all.json": func() any { return stackimage.ListResponse{} },
	"1.5/cloud/server/list_all.json":      func() any { return cloudserver.ListResponse{} },
}

func main() {
	src, dst := os.Args[1], os.Args[2]
	names, _ := filepath.Glob(filepath.Join(src, "*.json"))
	for _, n := range names {
		raw, err := os.ReadFile(n) //nolint:gosec // developer-supplied path
		if err != nil {
			fmt.Println("ERR", n, err)
			continue
		}
		var rec recorder.Recording
		if err := json.Unmarshal(raw, &rec); err != nil {
			fmt.Println("ERR decode", n, err)
			continue
		}
		out, err := recorder.Scrub([]byte(rec.Body))
		if err != nil {
			fmt.Println("ERR scrub", n, err)
			continue
		}
		if err := os.WriteFile(filepath.Join(dst, filepath.Base(n)), out, 0o600); err != nil {
			fmt.Println("ERR write", err)
			continue
		}
		mk, ok := types[rec.Endpoint]
		if !ok || !rec.OK {
			continue
		}
		missing, err := shapecheck.Undecoded([]byte(rec.Body), mk())
		if err != nil {
			fmt.Println("ERR shapecheck", err)
			continue
		}
		if len(missing) > 0 {
			fmt.Printf("%-38s UNDECODED: %s\n", rec.Endpoint, strings.Join(missing, ", "))
		}
	}
}
