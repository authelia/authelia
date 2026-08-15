package session

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSessionUsernameMustOnlyBeSetByAConstructor(t *testing.T) {
	root, err := findModuleRoot()

	require.NoError(t, err)

	var violations []string

	require.NoError(t, filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == "vendor" || name == "node_modules" || name == "web" {
				return fs.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		found, err := findUserSessionUsernameAssignments(path)
		if err != nil {
			return err
		}

		for _, violation := range found {
			rel, rerr := filepath.Rel(root, violation.file)
			if rerr != nil {
				rel = violation.file
			}

			violations = append(violations, fmt.Sprintf("%s:%d: %s.Username is assigned directly", rel, violation.line, violation.ident))
		}

		return nil
	}))

	assert.Empty(t, violations, "UserSession.Username must only be set by NewUserSession or Strategy.New, see the SECURITY NOTE on the field:\n%s", strings.Join(violations, "\n"))
}

type usernameAssignment struct {
	file  string
	line  int
	ident string
}

func findUserSessionUsernameAssignments(path string) (found []usernameAssignment, err error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	if isGenerated(file) {
		return nil, nil
	}

	local := file.Name.Name == "session" && strings.HasSuffix(filepath.Dir(path), filepath.Join("internal", "session"))

	if !local && !importsSession(file) {
		return nil, nil
	}

	idents := userSessionIdents(file, local)

	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for _, expr := range assign.Lhs {
			selector, ok := expr.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Username" {
				continue
			}

			ident, ok := selector.X.(*ast.Ident)
			if !ok || !idents[ident.Name] {
				continue
			}

			found = append(found, usernameAssignment{file: path, line: fset.Position(selector.Pos()).Line, ident: ident.Name})
		}

		return true
	})

	return found, nil
}

// isGenerated reports whether the file carries the conventional generated code marker. Generated marshalling code
// assigns the field when decoding a persisted session, which is deserialization rather than a change of identity.
func isGenerated(file *ast.File) bool {
	for _, group := range file.Comments {
		if group.Pos() > file.Package {
			break
		}

		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "// Code generated ") && strings.HasSuffix(comment.Text, " DO NOT EDIT.") {
				return true
			}
		}
	}

	return false
}

func importsSession(file *ast.File) bool {
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, `"`) == "github.com/authelia/authelia/v4/internal/session" {
			return true
		}
	}

	return false
}

func userSessionIdents(file *ast.File, local bool) (idents map[string]bool) {
	idents = map[string]bool{}

	record := func(names []*ast.Ident, expr ast.Expr) {
		if !isUserSessionType(expr, local) {
			return
		}

		for _, name := range names {
			idents[name.Name] = true
		}
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ValueSpec:
			record(node.Names, node.Type)
		case *ast.Field:
			record(node.Names, node.Type)
		case *ast.FuncDecl:
			if node.Recv != nil {
				for _, recv := range node.Recv.List {
					record(recv.Names, recv.Type)
				}
			}
		case *ast.AssignStmt:
			if node.Tok != token.DEFINE {
				return true
			}

			for i, lhs := range node.Lhs {
				name, ok := lhs.(*ast.Ident)
				if !ok || i >= len(node.Rhs) {
					continue
				}

				if isUserSessionValue(node.Rhs[i], local) {
					idents[name.Name] = true
				}
			}
		}

		return true
	})

	return idents
}

func isUserSessionType(expr ast.Expr, local bool) bool {
	switch node := expr.(type) {
	case *ast.StarExpr:
		return isUserSessionType(node.X, local)
	case *ast.Ident:
		return local && node.Name == "UserSession"
	case *ast.SelectorExpr:
		pkg, ok := node.X.(*ast.Ident)

		return ok && pkg.Name == "session" && node.Sel.Name == "UserSession"
	}

	return false
}

func isUserSessionValue(expr ast.Expr, local bool) bool {
	switch node := expr.(type) {
	case *ast.UnaryExpr:
		return node.Op == token.AND && isUserSessionValue(node.X, local)
	case *ast.CompositeLit:
		return isUserSessionType(node.Type, local)
	case *ast.StarExpr:
		return isUserSessionValue(node.X, local)
	case *ast.CallExpr:
		return isUserSessionConstructor(node.Fun)
	}

	return false
}

func isUserSessionConstructor(expr ast.Expr) bool {
	var name string

	switch node := expr.(type) {
	case *ast.Ident:
		name = node.Name
	case *ast.SelectorExpr:
		name = node.Sel.Name
	default:
		return false
	}

	switch name {
	case "NewUserSession", "NewDefaultUserSession", "NewDefault", "New", "GetSession", "Get":
		return true
	}

	return false
}

func findModuleRoot() (root string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate the module root from '%s'", dir)
		}

		dir = parent
	}
}
