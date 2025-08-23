package main

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func newDocsSEOCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seo",
		Short: "Generate doc seo tags",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newDocsSEOOpenIDConnectCmd())

	return cmd
}

func newDocsSEOOpenIDConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openid-connect",
		Short: "Generate doc openid connect integration guide seo tags",
		RunE:  docsSEOOpenIDConnectRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

//nolint:gocyclo
func docsSEOOpenIDConnectRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		pathDocsContent string
	)

	if pathDocsContent, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsContent); err != nil {
		return err
	}

	root := filepath.Join(pathDocsContent, "integration", "openid-connect", "clients")

	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "index.md" || filepath.Dir(path) == root {
			return nil
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return nil
		}

		frontmatterBytes := getFrontmatter(abs)

		if frontmatterBytes == nil {
			return nil
		}

		frontmatter := map[string]any{}

		if err = yaml.Unmarshal(frontmatterBytes, frontmatter); err != nil {
			return fmt.Errorf("error parsing frontmatter for file %s: %v", path, err)
		}

		var (
			ok                    bool
			raw                   any
			seo                   map[string]any
			seotitle, title, desc string
		)

		if raw, ok = frontmatter["seo"]; !ok {
			return fmt.Errorf("error parsing frontmatter for %s: seo is missing", path)
		}

		if seo, ok = raw.(map[string]any); !ok {
			return fmt.Errorf("error parsing frontmatter for %s: seo is not a map", path)
		}

		if raw, ok = seo["description"]; !ok {
			return fmt.Errorf("error parsing frontmatter for %s: seo description is missing", path)
		}

		if desc, ok = raw.(string); !ok {
			return fmt.Errorf("error parsing frontmatter for %s: seo description is not a string", path)
		}

		if raw, ok = frontmatter["title"]; !ok {
			return fmt.Errorf("error parsing frontmatter for %s: title is missing", path)
		}

		if title, ok = raw.(string); !ok {
			return fmt.Errorf("error parsing frontmatter for %s: title is not a string", path)
		}

		if raw, ok = seo["title"]; !ok || raw == "" {
			seotitle = fmt.Sprintf("%s | OpenID Connect 1.0 | Integration", title)
		}

		if len(seotitle) != 0 {
			replaceFrontMatter(abs, "", fmt.Sprintf(`  title: "%s"`, seotitle), "  title:")
		}

		expected := fmt.Sprintf("Step-by-step guide to configuring %s with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management.", title)

		if desc == expected {
			return nil
		}

		replaceFrontMatter(abs, "", fmt.Sprintf(`  description: "%s"`, expected), "  description:")

		return nil
	})
}
