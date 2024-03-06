package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/model"
)

func newGitHubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseGitHub,
		Short: "Generate GitHub files",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newGitHubIssueTemplatesCmd())

	return cmd
}

func newGitHubIssueTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseGitHubIssueTemplates,
		Short: "Generate GitHub issue templates",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newGitHubIssueTemplatesBugReportCmd(), newGitHubIssueTemplatesFeatureCmd())

	return cmd
}

func newGitHubIssueTemplatesFeatureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseGitHubIssueTemplatesFR,
		Short: "Generate GitHub feature request issue template",
		RunE:  cmdGitHubIssueTemplatesFeatureRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newGitHubIssueTemplatesBugReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseGitHubIssueTemplatesBR,
		Short: "Generate GitHub bug report issue template",
		RunE:  cmdGitHubIssueTemplatesBugReportRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func cmdGitHubIssueTemplatesFeatureRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		cwd, file, root  string
		tags, tagsFuture []string
		versions         int
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

	if versions, err = cmd.Flags().GetInt(cmdFlagVersionCount); err != nil {
		return err
	}

	if tags, err = getGitTags(cwd); err != nil {
		return err
	}

	var latest *model.SemanticVersion

	if latest, err = model.NewSemanticVersion(tags[0]); err != nil {
		return fmt.Errorf("error extracting latest minor version from tag: %w", err)
	}

	for i := 0; i < versions; i++ {
		tagsFuture = append(tagsFuture, fmt.Sprintf("v%s", model.SemanticVersion{Major: latest.Major, Minor: latest.Minor + i + 1}.String()))
	}

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
		cwd, file, dirRoot string
		versions           int

		tags []string
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

	if versions, err = cmd.Flags().GetInt(cmdFlagVersionCount); err != nil {
		return err
	}

	if tags, err = getGitTags(cwd); err != nil {
		return err
	}

	var latest, version *model.SemanticVersion

	if latest, err = model.NewSemanticVersion(tags[0]); err != nil {
		return fmt.Errorf("error extracting latest minor version from tag: %w", err)
	}

	minimum := latest.Copy()

	minimum.Patch = 0
	minimum.Minor -= versions

	var tagsRecent []string

	for _, tag := range tags {
		if len(tag) == 0 {
			continue
		}

		if version, err = model.NewSemanticVersion(tag); err != nil {
			return fmt.Errorf("error extracting minor version from tag: %w", err)
		}

		if !version.IsStable() {
			continue
		}

		if version.GreaterThanOrEqual(minimum) {
			tagsRecent = append(tagsRecent, tag)
		}
	}

	var (
		f *os.File
	)

	fullPath := filepath.Join(dirRoot, file)

	if f, err = os.Create(fullPath); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPath, err)
	}

	data := &tmplIssueTemplateData{
		Labels:   []string{labelTypeBugUnconfirmed.String(), labelStatusNeedsTriage.String(), labelPriorityNormal.String()},
		Versions: tagsRecent,
		Proxies:  []string{"Caddy", "Traefik", "Envoy", "Istio", "NGINX", "SWAG", "NGINX Proxy Manager", "HAProxy"},
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
