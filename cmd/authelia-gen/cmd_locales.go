package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
		pathWebI18NIndex      string
		pathDocsDataLanguages string
	)

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathLocales, err = cmd.Flags().GetString(cmdFlagDirLocales); err != nil {
		return err
	}

	if pathWebI18NIndex, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagWeb, cmdFlagFileWebI18N); err != nil {
		return err
	}

	if pathDocsDataLanguages, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsData, cmdFlagDocsDataLanguages); err != nil {
		return err
	}

	var data *utils.Languages

	if data, err = utils.GetDirectoryLanguages(filepath.Join(root, pathLocales)); err != nil {
		return err
	}

	fWriteWebI18NIndex := func(path string, data *utils.Languages) (err error) {
		var (
			f *os.File
		)

		if f, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600); err != nil {
			return fmt.Errorf("failed to write file '%s': %w", path, err)
		}

		defer func() {
			_ = f.Close()
		}()

		type valueType struct {
			Languages    map[string][]string
			LanguageKeys []string

			Data *utils.Languages
		}

		values := &valueType{
			Languages:    map[string][]string{},
			LanguageKeys: []string{},
			Data:         data,
		}

		for _, language := range data.Languages {
			values.LanguageKeys = append(values.LanguageKeys, language.Locale)
			values.Languages[language.Locale] = language.Fallbacks
		}

		values.LanguageKeys = append(values.LanguageKeys, "default")
		values.Languages["default"] = []string{values.Data.Defaults.Language.Locale}

		sort.Strings(values.LanguageKeys)

		if err = tmplWebI18NIndex.Execute(f, values); err != nil {
			return err
		}

		return nil
	}

	fWriteDocsDataLanguages := func(path string, data *utils.Languages) (err error) {
		var (
			f *os.File
		)

		if f, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600); err != nil {
			return fmt.Errorf("failed to write file '%s': %w", path, err)
		}

		defer func() {
			_ = f.Close()
		}()

		encoder := json.NewEncoder(f)

		encoder.SetIndent("", "    ")

		if err = encoder.Encode(data); err != nil {
			return fmt.Errorf("failed to encode json data: %w", err)
		}

		return nil
	}

	if err = fWriteWebI18NIndex(pathWebI18NIndex, data); err != nil {
		return err
	}

	if err = fWriteDocsDataLanguages(pathDocsDataLanguages, data); err != nil {
		return err
	}

	return nil
}
