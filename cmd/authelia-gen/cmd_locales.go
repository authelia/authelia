package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"github.com/authelia/authelia/v4/internal/utils"
)

func newLocalesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "locales",
		Short:             "Generate locales files",
		RunE:              localesRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func localesRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		root, pathLocales                       string
		pathWebI18NIndex, pathDocsDataLanguages string
	)

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathLocales, err = cmd.Flags().GetString(cmdFlagDirLocales); err != nil {
		return err
	}

	if pathWebI18NIndex, err = cmd.Flags().GetString(cmdFlagFileWebI18N); err != nil {
		return err
	}

	if pathDocsDataLanguages, err = cmd.Flags().GetString(cmdFlagDocsDataLanguages); err != nil {
		return err
	}

	data, err := getLanguages(filepath.Join(root, pathLocales))
	if err != nil {
		return err
	}

	fullPathWebI18NIndex := filepath.Join(root, pathWebI18NIndex)

	var (
		f        *os.File
		dataJSON []byte
	)

	if f, err = os.Create(fullPathWebI18NIndex); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPathWebI18NIndex, err)
	}

	if err = tmplWebI18NIndex.Execute(f, data); err != nil {
		return err
	}

	if dataJSON, err = json.Marshal(data); err != nil {
		return err
	}

	fullPathDocsDataLanguages := filepath.Join(root, pathDocsDataLanguages)

	if err = os.WriteFile(fullPathDocsDataLanguages, dataJSON, 0600); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", fullPathDocsDataLanguages, err)
	}

	return nil
}

func getLanguages(dir string) (languages *Languages, err error) {
	var locales []string

	languages = &Languages{
		DefaultLocale:    localeDefault,
		DefaultNamespace: localeNamespaceDefault,
	}

	if err = filepath.Walk(dir, func(path string, info fs.FileInfo, errWalk error) (err error) {
		if errWalk != nil {
			return errWalk
		}

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

		var tag language.Tag

		if tag, err = language.Parse(locale); err != nil {
			fmt.Println(err)
		}

		l := Language{
			Display:    display.English.Tags().Name(tag),
			Locale:     locale,
			Namespaces: []string{ns},
		}

		languages.Languages = append(languages.Languages, l)

		locales = append(locales, locale)

		return nil
	}); err != nil {
		return nil, err
	}

	sort.Slice(languages.Languages, func(i, j int) bool {
		return languages.Languages[i].Locale == localeDefault || languages.Languages[i].Locale < languages.Languages[j].Locale
	})

	for i, l := range languages.Languages {
		parts := strings.SplitN(l.Locale, "-", 2)
		if len(parts) == 2 && utils.IsStringInSlice(parts[0], locales) {
			languages.Languages[i].Fallbacks = append(languages.Languages[i].Fallbacks, parts[0])
		}

		languages.Languages[i].Fallbacks = append(languages.Languages[i].Fallbacks, languages.DefaultLocale)
	}

	return languages, nil
}
