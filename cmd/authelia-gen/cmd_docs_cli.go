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
		Use:   "cli",
		Short: "Generate CLI docs",
		RunE:  docsCLIRunE,
	}

	cmd.Flags().StringP("directory", "d", "./docs/content/en/reference/cli", "The directory to store the markdown in")

	return cmd
}

func docsCLIRunE(cmd *cobra.Command, args []string) (err error) {
	var root string

	if root, err = cmd.Flags().GetString("directory"); err != nil {
		return err
	}

	if err = os.MkdirAll(root, 0775); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	if err = genCLIDoc(commands.NewRootCmd(), filepath.Join(root, "authelia")); err != nil {
		return err
	}

	if err = genCLIDocWriteIndex(root, "authelia"); err != nil {
		return err
	}

	if err = genCLIDoc(cmdscripts.NewRootCmd(), filepath.Join(root, "authelia-scripts")); err != nil {
		return err
	}

	if err = genCLIDocWriteIndex(root, "authelia-scripts"); err != nil {
		return err
	}

	if err = genCLIDoc(newRootCmd(), filepath.Join(root, "authelia-gen")); err != nil {
		return err
	}

	if err = genCLIDocWriteIndex(root, "authelia-gen"); err != nil {
		return err
	}

	return nil
}

func genCLIDoc(cmd *cobra.Command, path string) (err error) {
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

	weight := 900

	if name == "authelia" {
		weight = 320
	}

	_, err = fmt.Fprintf(f, indexDocs, name, now.Format(dateFmtYAML), "cli-"+name, weight)

	return err
}

func prepend(input string) string {
	now := time.Now()

	pathz := strings.Split(strings.Replace(input, ".md", "", 1), "\\")
	parts := strings.Split(pathz[len(pathz)-1], "_")

	cmd := parts[0]

	args := strings.Join(parts, " ")

	weight := 330
	if len(parts) == 1 {
		weight = 320
	}

	return fmt.Sprintf(prefixDocs, args, fmt.Sprintf("Reference for the %s command.", args), "", now.Format(dateFmtYAML), "cli-"+cmd, weight)
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
menu:
  reference:
    parent: "cli"
    identifier: "%s"
weight: %d
toc: true
---
`

const prefixDocs = `---
title: "%s"
description: "%s"
lead: "%s"
date: %s
draft: false
images: []
menu:
  reference:
    parent: "%s"
weight: %d
toc: true
---

`
