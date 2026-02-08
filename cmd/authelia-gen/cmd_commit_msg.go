package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// CommitMessageTmpl is a template data structure which is used to generate files with commit message information.
type CommitMessageTmpl struct {
	Scopes ScopesTmpl
	Types  TypesTmpl
}

// TypesTmpl is a template data structure which is used to generate files with commit message types.
type TypesTmpl struct {
	List    []string
	Details []NameDescriptionTmpl
}

// ScopesTmpl is a template data structure which is used to generate files with commit message scopes.
type ScopesTmpl struct {
	All      []string
	Packages []string
	Extra    []NameDescriptionTmpl
}

// NameDescriptionTmpl is a template item which includes a name, description and list of scopes.
type NameDescriptionTmpl struct {
	Name        string
	Description string
	Scopes      []string
}

func newCommitLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseCommitLint,
		Short: "Generate commit lint files",
		RunE:  commitLintRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

var commitScopesExtra = []NameDescriptionTmpl{
	{"api", "used for changes that change the openapi specification", nil},
	{"cmd", "used for changes to the `%s` top level binaries", nil},
	{"web", "used for changes to the React based frontend", nil},
}

var commitTypes = []NameDescriptionTmpl{
	{"build", "Changes that affect the build system or external dependencies", []string{"bundler", "deps", "docker", "go", "npm"}},
	{"ci", "Changes to our CI configuration files and scripts", []string{"autheliabot", "buildkite", "codecov", "lefthook", "golangci-lint", "renovate", "reviewdog"}},
	{"docs", "Documentation only changes", nil},
	{"feat", "A new feature", nil},
	{"fix", "A bug fix", nil},
	{"i18n", "Updating translations or internationalization settings", nil},
	{"perf", "A code change that improves performance", nil},
	{"refactor", "A code change that neither fixes a bug nor adds a feature", nil},
	{"release", "Releasing a new version of Authelia", nil},
	{"test", "Adding missing tests or correcting existing tests", nil},
}

var commitTypesExtra = []string{"revert"}

func getGoPackages(dir string) (pkgs []string, err error) {
	var (
		entries    []os.DirEntry
		entriesSub []os.DirEntry
	)

	if entries, err = os.ReadDir(dir); err != nil {
		return nil, fmt.Errorf("failed to detect go packages in directory '%s': %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if entriesSub, err = os.ReadDir(filepath.Join(dir, entry.Name())); err != nil {
			continue
		}

		for _, entrySub := range entriesSub {
			if entrySub.IsDir() {
				continue
			}

			if strings.HasSuffix(entrySub.Name(), ".go") {
				pkgs = append(pkgs, entry.Name())
				break
			}
		}
	}

	return pkgs, nil
}

func commitLintRunE(cmd *cobra.Command, args []string) (err error) {
	var root, pathCommitLintConfig, pathDocsCommitMessageGuidelines string

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathCommitLintConfig, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagWeb, cmdFlagFileConfigCommitLint); err != nil {
		return err
	}

	if pathDocsCommitMessageGuidelines, err = cmd.Flags().GetString(cmdFlagFileDocsCommitMsgGuidelines); err != nil {
		return err
	}

	data := &CommitMessageTmpl{
		Scopes: ScopesTmpl{
			All:      []string{},
			Packages: []string{},
			Extra:    []NameDescriptionTmpl{},
		},
		Types: TypesTmpl{
			List:    []string{},
			Details: []NameDescriptionTmpl{},
		},
	}

	var (
		cmds []string
		pkgs []string
	)

	if cmds, err = getGoPackages(filepath.Join(root, subPathCmd)); err != nil {
		return err
	}

	if pkgs, err = getGoPackages(filepath.Join(root, subPathInternal)); err != nil {
		return err
	}

	data.Scopes.All = append(data.Scopes.All, pkgs...)
	data.Scopes.Packages = append(data.Scopes.Packages, pkgs...)

	for _, scope := range commitScopesExtra {
		switch scope.Name {
		case subPathCmd:
			data.Scopes.Extra = append(data.Scopes.Extra, NameDescriptionTmpl{Name: scope.Name, Description: fmt.Sprintf(scope.Description, strings.Join(cmds, "|"))})
		default:
			data.Scopes.Extra = append(data.Scopes.Extra, scope)
		}

		data.Scopes.All = append(data.Scopes.All, scope.Name)
	}

	for _, cType := range commitTypes {
		data.Types.List = append(data.Types.List, cType.Name)
		data.Types.Details = append(data.Types.Details, cType)

		data.Scopes.All = append(data.Scopes.All, cType.Scopes...)
	}

	data.Types.List = append(data.Types.List, commitTypesExtra...)

	sort.Slice(data.Scopes.All, func(i, j int) bool {
		return data.Scopes.All[i] < data.Scopes.All[j]
	})

	sort.Slice(data.Scopes.Packages, func(i, j int) bool {
		return data.Scopes.Packages[i] < data.Scopes.Packages[j]
	})

	sort.Slice(data.Scopes.Extra, func(i, j int) bool {
		return data.Scopes.Extra[i].Name < data.Scopes.Extra[j].Name
	})

	sort.Slice(data.Types.List, func(i, j int) bool {
		return data.Types.List[i] < data.Types.List[j]
	})

	sort.Slice(data.Types.Details, func(i, j int) bool {
		return data.Types.Details[i].Name < data.Types.Details[j].Name
	})

	var f *os.File

	fullPathCommitLintConfig := filepath.Join(root, pathCommitLintConfig)

	if f, err = os.Create(fullPathCommitLintConfig); err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", fullPathCommitLintConfig, err)
	}

	if err = tmplDotCommitLintRC.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write output file '%s': %w", fullPathCommitLintConfig, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close output file '%s': %w", fullPathCommitLintConfig, err)
	}

	fullPathDocsCommitMessageGuidelines := filepath.Join(root, pathDocsCommitMessageGuidelines)

	if f, err = os.Create(fullPathDocsCommitMessageGuidelines); err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", fullPathDocsCommitMessageGuidelines, err)
	}

	if err = tmplDocsCommitMessageGuidelines.Execute(f, data); err != nil {
		return fmt.Errorf("failed to write output file '%s': %w", fullPathDocsCommitMessageGuidelines, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close output file '%s': %w", fullPathDocsCommitMessageGuidelines, err)
	}

	return nil
}
