package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	cmdscripts "github.com/authelia/authelia/v4/cmd/authelia-scripts/cmd"
	"github.com/authelia/authelia/v4/internal/commands"
)

func newDocsCLICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocsCLI,
		Short: "Generate CLI docs",
		RunE:  docsCLIRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func docsCLIRunE(cmd *cobra.Command, args []string) (err error) {
	var outputPath string

	if outputPath, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsContent, cmdFlagDocsCLIReference); err != nil {
		return err
	}

	if err = os.MkdirAll(outputPath, 0775); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if err = genCLIDoc(commands.NewRootCmd(), filepath.Join(outputPath, "authelia")); err != nil {
		return err
	}

	if err = genCLIDocWriteIndex(outputPath, "authelia"); err != nil {
		return err
	}

	if err = genCLIDoc(cmdscripts.NewRootCmd(), filepath.Join(outputPath, "authelia-scripts")); err != nil {
		return err
	}

	if err = genCLIDocWriteIndex(outputPath, "authelia-scripts"); err != nil {
		return err
	}

	if err = genCLIDoc(newRootCmd(), filepath.Join(outputPath, cmdUseRoot)); err != nil {
		return err
	}

	if err = genCLIDocWriteIndex(outputPath, cmdUseRoot); err != nil {
		return err
	}

	return nil
}

func genCLIDoc(cmd *cobra.Command, path string) (err error) {
	if _, err = os.Stat(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil || !os.IsNotExist(err) {
		if err = os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove docs: %w", err)
		}
	}

	if err = os.Mkdir(path, 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if err = doc.GenMarkdownTreeCustom(cmd, path, prepend, linker); err != nil {
		return err
	}

	return nil
}

func genCLIDocWriteIndex(path, name string) (err error) {
	now := time.Now()

	f, err := os.Create(filepath.Join(path, name, "_index.md"))
	if err != nil {
		return err
	}

	weight := genCLIDocCmdToWeight(name)

	_, err = fmt.Fprintf(f, indexDocs, name, now.Format(dateFmtYAML), weight)

	return err
}

func prepend(input string) string {
	now := time.Now()

	_, filename := filepath.Split(strings.Replace(input, ".md", "", 1))

	parts := strings.Split(filename, "_")

	args := strings.Join(parts, " ")

	weight := genCLIDocCmdToWeight(parts[0])

	if len(parts) != 1 {
		weight += 5
	}

	return fmt.Sprintf(prefixDocs, args, fmt.Sprintf("Reference for the %s command.", args), "", now.Format(dateFmtYAML), weight)
}

func genCLIDocCmdToWeight(cmd string) int {
	switch cmd {
	case "authelia":
		return 900
	case "authelia-gen":
		return 910
	case "authelia-scripts":
		return 920
	default:
		return 990
	}
}

func linker(input string) string {
	return input
}

const indexDocs = `---
title: "%s"
description: ""
lead: ""
date: %s
draft: false
images: []
sidebar:
  collapsed: true
weight: %d
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---
`

const prefixDocs = `---
title: "%s"
description: "%s"
lead: "%s"
date: %s
draft: false
images: []
weight: %d
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

`
