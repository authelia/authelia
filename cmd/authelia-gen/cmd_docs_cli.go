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

	if err = genCLIDoc(cmdscripts.NewRootCmd(), filepath.Join(root, "authelia-scripts")); err != nil {
		return err
	}

	if err = genCLIDoc(newRootCmd(), filepath.Join(root, "authelia-gen")); err != nil {
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

const prefixDocs = `---
title: "%s"
description: "%s"
lead: "%s"
date: 2022-05-30T06:42:39+10:00
lastmod: %s
draft: false
images: []
menu:
  reference:
    parent: "%s"
weight: %d
toc: true
---

`

func prepend(input string) string {
	fmt.Printf("prepending: %s\n", input)

	now := time.Now()

	pathz := strings.Split(strings.Replace(input, ".md", "", 1), "\\")
	parts := strings.Split(pathz[len(pathz)-1], "_")

	cmd := parts[0]

	args := strings.Join(parts, " ")

	return fmt.Sprintf(prefixDocs, args, fmt.Sprintf("Reference for the %s command.", args), "", now.Format("2006-01-02T15:04:05-07:00"), "cli-"+cmd, 130)
}

func linker(input string) string {
	fmt.Printf("linking: %s\n", input)
	return input
}
