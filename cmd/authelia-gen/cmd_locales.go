package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newLocalesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locales",
		Short: "Generate locales files",
		RunE:  runE,
	}

	cmd.AddCommand(newGitHubIssueTemplatesCmd())

	return cmd
}

func runE(cmd *cobra.Command, args []string) (err error) {
	data, err := getLanguages("./internal/server/locales/")
	if err != nil {
		return err
	}

	file := "./web/src/i18n/index.ts"

	var (
		f        *os.File
		dataJSON []byte
	)

	if f, err = os.Create(file); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", file, err)
	}

	if err = tmplWebI18NIndex.Execute(f, data); err != nil {
		return err
	}

	if dataJSON, err = json.Marshal(data); err != nil {
		return err
	}

	if err = os.WriteFile("./docs/data/languages.json", dataJSON, 0644); err != nil {
		return err
	}

	return nil
}

func getLanguages(dir string) (languages *Languages, err error) {
	var locales []string
	languages = &Languages{
		DefaultLocale:    defaultLocale,
		DefaultNamespace: defaultNamespace,
	}

	if err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		nameLower := strings.ToLower(info.Name())
		ext := filepath.Ext(nameLower)
		ns := strings.Replace(nameLower, ext, "", 1)

		if ext != ".json" {
			return nil
		}

		if !utils.IsStringInSlice(ns, languages.Namespaces) {
			languages.Namespaces = append(languages.Namespaces, ns)
		}

		fdir, _ := filepath.Split(path)

		locale := filepath.Base(fdir)

		if utils.IsStringInSlice(locale, locales) {
			for i, l := range languages.Languages {
				if l.Locale == locale {
					if utils.IsStringInSlice(ns, languages.Languages[i].Namespaces) {
						break
					}

					languages.Languages[i].Namespaces = append(languages.Languages[i].Namespaces, ns)
					break
				}
			}

			return nil
		}

		lang, err := language.Parse(locale)
		if err != nil {
			fmt.Println(err)
		}

		l := Language{
			Display:    display.English.Tags().Name(lang),
			Locale:     locale,
			Namespaces: []string{ns},
		}

		languages.Languages = append(languages.Languages, l)

		locales = append(locales, locale)

		return nil
	}); err != nil {
		return nil, err
	}

	for i, l := range languages.Languages {
		parts := strings.SplitN(l.Locale, "-", 2)
		if len(parts) == 2 && utils.IsStringInSlice(parts[0], locales) {
			languages.Languages[i].Fallbacks = append(languages.Languages[i].Fallbacks, parts[0])
		}

		languages.Languages[i].Fallbacks = append(languages.Languages[i].Fallbacks, languages.DefaultLocale)
	}

	return languages, nil
}
