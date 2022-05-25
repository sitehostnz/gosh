//go:build tools
// +build tools

package gosh

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/ory/go-acc"
)
