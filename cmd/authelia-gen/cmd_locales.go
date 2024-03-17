package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newLocalesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseLocales,
		Short: "Generate locales files",
		RunE:  localesRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func localesRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		root, pathLocales     string
		pathDocsDataLanguages string
	)

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathLocales, err = cmd.Flags().GetString(cmdFlagDirLocales); err != nil {
		return err
	}

	// if pathWebI18NIndex, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagWeb, cmdFlagFileWebI18N); err != nil {
	// 	return err
	// }.

	if pathDocsDataLanguages, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsData, cmdFlagDocsDataLanguages); err != nil {
		return err
	}

	data, err := utils.GetCustomLanguages(filepath.Join(root, pathLocales))
	if err != nil {
		return err
	}

	// fullPathWebI18NIndex := filepath.Join(root, pathWebI18NIndex).

	var (
		f *os.File
	)

	// if f, err = os.Create(fullPathWebI18NIndex); err != nil {
	// 	return fmt.Errorf("failed to create file '%s': %w", fullPathWebI18NIndex, err)
	// }.

	// if err = tmplWebI18NIndex.Execute(f, data); err != nil {
	// 	return err
	// }.

	// _ = f.Close().

	fullPathDocsDataLanguages := filepath.Join(root, pathDocsDataLanguages)

	if f, err = os.OpenFile(fullPathDocsDataLanguages, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", fullPathDocsDataLanguages, err)
	}

	encoder := json.NewEncoder(f)

	encoder.SetIndent("", "    ")

	if err = encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode json data: %w", err)
	}

	return nil
}
