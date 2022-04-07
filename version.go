package gosh

import (
	"fmt"
	"runtime/debug"
)

var (
	Version          = "dev"
	DefaultUserAgent string
)

const packagePath = "github.com/gonzariosm/sitehost-go"

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == packagePath {
				if dep.Replace != nil {
					Version = dep.Replace.Version
				}
				Version = dep.Version
				break
			}
		}
	}

	DefaultUserAgent = fmt.Sprintf("sitehost/%s %s", Version, packagePath)
}
