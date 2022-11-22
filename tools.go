//go:build tools
// +build tools

// Package gosh provides the functions to work with SiteHost API.
package gosh

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/ory/go-acc"
)
