package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const (
	cmdUseRelease = "release"

	releaseTypeMajor = "major"
	releaseTypeMinor = "minor"
	releaseTypePatch = "patch"
)

func newReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseRelease,
		Short: "Prepare a release by updating version references, accepts one argument of either major, minor, or patch",
		Args:  cobra.ExactArgs(1),
		RunE:  releaseRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func releaseRunE(cmd *cobra.Command, args []string) (err error) {
	releaseType := args[0]

	switch releaseType {
	case releaseTypeMajor, releaseTypeMinor, releaseTypePatch:
		break
	default:
		return fmt.Errorf("invalid release type '%s': must be one of '%s', '%s', or '%s'", releaseType, releaseTypeMajor, releaseTypeMinor, releaseTypePatch)
	}

	current, err := readVersion(cmd)
	if err != nil {
		return fmt.Errorf("failed to read current version: %w", err)
	}

	var next string

	switch releaseType {
	case releaseTypeMajor:
		next = current.NextMajor().String()
	case releaseTypeMinor:
		next = current.NextMinor().String()
	case releaseTypePatch:
		next = current.NextPatch().String()
	}

	var root string

	if root, err = getPFlagPath(cmd.Flags(), cmdFlagRoot); err != nil {
		return err
	}

	if err = releaseUpdateWebPackageJSON(cmd, next); err != nil {
		return fmt.Errorf("failed to update web/package.json: %w", err)
	}

	rootCmd.SetArgs([]string{"--exclude", "docs.cli,docs.date"})

	if err = rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to run authelia-gen: %w", err)
	}

	var bugReportPath string

	if bugReportPath, err = cmd.Flags().GetString(cmdFlagBugReport); err != nil {
		return err
	}

	if err = releaseUpdateBugReport(filepath.Join(root, bugReportPath), next); err != nil {
		return fmt.Errorf("failed to update bug report template: %w", err)
	}

	var docsContentPath string

	if docsContentPath, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsContent); err != nil {
		return err
	}

	if err = releaseUpdateOIDCClientDocs(docsContentPath, current.String(), next); err != nil {
		return fmt.Errorf("failed to update OIDC client docs: %w", err)
	}

	return nil
}

func releaseUpdateWebPackageJSON(cmd *cobra.Command, version string) (err error) {
	var pathPackageJSON string

	if pathPackageJSON, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagWeb, cmdFlagFileWebPackage); err != nil {
		return err
	}

	var data []byte

	if data, err = os.ReadFile(pathPackageJSON); err != nil {
		return err
	}

	re := regexp.MustCompile(`("version"\s*:\s*")([^"]+)(")`)

	data = re.ReplaceAll(data, []byte(fmt.Sprintf("${1}%s${3}", version)))

	var f *os.File

	if f, err = os.Create(pathPackageJSON); err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)

	return err
}

func releaseUpdateBugReport(path, version string) (err error) {
	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var buf bytes.Buffer

	newOption := fmt.Sprintf("        - 'v%s'", version)

	optionsFound := false
	inserted := false

	for scanner.Scan() {
		line := scanner.Text()

		buf.WriteString(line)
		buf.WriteByte('\n')

		if !inserted && !optionsFound && strings.TrimSpace(line) == "options:" {
			optionsFound = true

			continue
		}

		if !inserted && optionsFound && strings.HasPrefix(strings.TrimSpace(line), "- 'v") {
			content := buf.String()
			content = content[:len(content)-len(line)-1]

			buf.Reset()
			buf.WriteString(content)
			buf.WriteString(newOption)
			buf.WriteByte('\n')
			buf.WriteString(line)
			buf.WriteByte('\n')

			inserted = true
		}
	}

	var f *os.File

	if f, err = os.Create(path); err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(buf.Bytes())

	return err
}

func releaseUpdateOIDCClientDocs(docsContentPath, currentVersion, nextVersion string) (err error) {
	glob := filepath.Join(docsContentPath, "integration", "openid-connect", "clients", "*", "index.md")

	var matches []string

	if matches, err = filepath.Glob(glob); err != nil {
		return err
	}

	currentPattern := fmt.Sprintf("[v%s](https://github.com/authelia/authelia/releases/tag/v%s)", currentVersion, currentVersion)
	nextReplacement := fmt.Sprintf("[v%s](https://github.com/authelia/authelia/releases/tag/v%s)", nextVersion, nextVersion)

	for _, match := range matches {
		if err = releaseUpdateFile(match, currentPattern, nextReplacement); err != nil {
			return fmt.Errorf("failed to update '%s': %w", match, err)
		}
	}

	return nil
}

func releaseUpdateFile(path, old, replacement string) (err error) {
	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return err
	}

	content := string(data)

	if !strings.Contains(content, old) {
		return nil
	}

	content = strings.ReplaceAll(content, old, replacement)

	var f *os.File

	if f, err = os.Create(path); err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(content)

	return err
}
