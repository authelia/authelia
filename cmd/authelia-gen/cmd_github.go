package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newGitHubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "github",
		Short:             "Generate GitHub files",
		RunE:              rootSubCommandsRunE,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newGitHubIssueTemplatesCmd())

	return cmd
}

func newGitHubIssueTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "issue-templates",
		Short:             "Generate GitHub issue templates",
		RunE:              rootSubCommandsRunE,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newGitHubIssueTemplatesBugReportCmd(), newGitHubIssueTemplatesFeatureCmd())

	return cmd
}

func newGitHubIssueTemplatesFeatureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "feature-request",
		Short:             "Generate GitHub feature request issue template",
		RunE:              cmdGitHubIssueTemplatesFeatureRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newGitHubIssueTemplatesBugReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "bug-report",
		Short:             "Generate GitHub bug report issue template",
		RunE:              cmdGitHubIssueTemplatesBugReportRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdGitHubIssueTemplatesFeatureRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		cwd, file, root                                 string
		tags, tagsFuture                                []string
		latestMajor, latestMinor, latestPatch, versions int
	)

	if cwd, err = cmd.Flags().GetString(cmdFlagCwd); err != nil {
		return err
	}

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if file, err = cmd.Flags().GetString(cmdFlagFeatureRequest); err != nil {
		return err
	}

	if versions, err = cmd.Flags().GetInt("versions"); err != nil {
		return err
	}

	if tags, err = getGitTags(cwd); err != nil {
		return err
	}

	latest := tags[0]

	if _, err = fmt.Sscanf(latest, "v%d.%d.%d", &latestMajor, &latestMinor, &latestPatch); err != nil {
		return fmt.Errorf("error occurred parsing version as semver: %w", err)
	}

	var (
		minor int
	)

	for minor = latestMinor + 1; minor < latestMinor+versions; minor++ {
		tagsFuture = append(tagsFuture, fmt.Sprintf("v%d.%d.0", latestMajor, minor))
	}

	tagsFuture = append(tagsFuture, fmt.Sprintf("v%d.0.0", latestMajor+1))

	var (
		f *os.File
	)

	fullPath := filepath.Join(root, file)

	if f, err = os.Create(fullPath); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPath, err)
	}

	data := &tmplIssueTemplateData{
		Labels:   []string{labelTypeFeature.String(), labelStatusNeedsDesign.String(), labelPriorityNormal.String()},
		Versions: tagsFuture,
	}

	if err = tmplIssueTemplateFeature.Execute(f, data); err != nil {
		return err
	}

	return nil
}

func cmdGitHubIssueTemplatesBugReportRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		cwd, file, dirRoot    string
		tags, tagsRecent      []string
		latestMinor, versions int
	)

	if cwd, err = cmd.Flags().GetString(cmdFlagCwd); err != nil {
		return err
	}

	if dirRoot, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if file, err = cmd.Flags().GetString(cmdFlagBugReport); err != nil {
		return err
	}
	if versions, err = cmd.Flags().GetInt("versions"); err != nil {
		return err
	}

	if tags, err = getGitTags(cwd); err != nil {
		return err
	}

	latest := tags[0]

	latestParts := strings.Split(latest, ".")

	if len(latestParts) < 2 {
		return fmt.Errorf("error extracting latest minor version from tag: %s does not appear to be a semver", latest)
	}

	if latestMinor, err = strconv.Atoi(latestParts[1]); err != nil {
		return fmt.Errorf("error extracting latest minor version from tag: %w", err)
	}

	var (
		parts []string
		minor int
	)

	for _, tag := range tags {
		if parts = strings.Split(tag, "."); len(parts) < 2 {
			return fmt.Errorf("error extracting minor version from tag: %s does not appear to be a semver", tag)
		}

		if minor, err = strconv.Atoi(parts[1]); err != nil {
			return fmt.Errorf("error extracting minor version from tag: %w", err)
		}

		if minor < latestMinor-versions {
			break
		}

		tagsRecent = append(tagsRecent, tag)
	}

	var (
		f *os.File
	)

	fullPath := filepath.Join(dirRoot, file)

	if f, err = os.Create(fullPath); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPath, err)
	}

	data := &tmplIssueTemplateData{
		Labels:   []string{labelTypePotentialBug.String(), labelStatusNeedsTriage.String(), labelPriorityNormal.String()},
		Versions: tagsRecent,
		Proxies:  []string{"Caddy", "Traefik", "NGINX", "SWAG", "NGINX Proxy Manager", "HAProxy"},
	}

	if err = tmplGitHubIssueTemplateBug.Execute(f, data); err != nil {
		return err
	}

	return nil
}

func getGitTags(cwd string) (tags []string, err error) {
	var (
		args       []string
		tagsOutput []byte
	)

	if len(cwd) != 0 {
		args = append(args, "-C", cwd)
	}

	args = append(args, "tag", "--sort=-creatordate")

	cmd := exec.Command("git", args...)

	if tagsOutput, err = cmd.Output(); err != nil {
		return nil, err
	}

	return strings.Split(string(tagsOutput), "\n"), nil
}
