//go:build tools
// +build tools

// Package toolspkgs exists only to pin CLI/development tools in go.mod.
// It is never compiled in normal builds.
package toolspkgs

import (
	_ "github.com/cespare/reflex"
	_ "github.com/go-delve/delve/cmd/dlv"
)
