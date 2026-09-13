// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEveryNotifierSendSiteShouldEmitAnEvent(t *testing.T) {
	for _, dir := range []string{"../handlers", "../middlewares"} {
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}

			path := filepath.Join(dir, entry.Name())

			fset := token.NewFileSet()

			file, err := parser.ParseFile(fset, path, nil, 0)
			require.NoError(t, err)

			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					return true
				}

				var sends, emits bool

				ast.Inspect(fn.Body, func(inner ast.Node) bool {
					sel, ok := inner.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					switch sel.Sel.Name {
					case "Send":
						if exprHasIdent(sel.X, "Notifier") {
							sends = true
						}
					case "Emit":
						emits = true
					}

					return true
				})

				assert.False(t, sends && !emits,
					"%s: function %s calls Notifier.Send without emitting an event; every notification must also produce a webhook event",
					path, fn.Name.Name)

				return true
			})
		}
	}
}

func exprHasIdent(expr ast.Expr, name string) (found bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == name {
			found = true
		}

		return !found
	})

	return found
}
