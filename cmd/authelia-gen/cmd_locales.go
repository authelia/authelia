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
		Use:   cmdUseLocales,
		Short: "Generate locales files",
		RunE:  localesRunE,

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

	if pathWebI18NIndex, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagWeb, cmdFlagFileWebI18N); err != nil {
		return err
	}

	if pathDocsDataLanguages, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsData, cmdFlagDocsDataLanguages); err != nil {
		return err
	}

	data, err := getLanguages(filepath.Join(root, pathLocales))
	if err != nil {
		return err
	}

	fullPathWebI18NIndex := filepath.Join(root, pathWebI18NIndex)

	var (
		f *os.File
	)

	if f, err = os.Create(fullPathWebI18NIndex); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPathWebI18NIndex, err)
	}

	if err = tmplWebI18NIndex.Execute(f, data); err != nil {
		return err
	}

	_ = f.Close()

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

//nolint:gocyclo
func getLanguages(dir string) (languages *Languages, err error) {
	//nolint:prealloc
	var locales []string

	languages = &Languages{
		Defaults: DefaultsLanguages{
			Namespace: localeNamespaceDefault,
		},
	}

	var defaultTag language.Tag

	if defaultTag, err = language.Parse(localeDefault); err != nil {
		return nil, fmt.Errorf("failed to parse default language: %w", err)
	}

	languages.Defaults.Language = Language{
		Display: display.English.Tags().Name(defaultTag),
		Locale:  localeDefault,
	}

	if err = filepath.Walk(dir, func(path string, info fs.FileInfo, errWalk error) (err error) {
		if errWalk != nil {
			return errWalk
		}

		nameLower := strings.ToLower(info.Name())
		ext := filepath.Ext(nameLower)
		ns := strings.Replace(nameLower, ext, "", 1)

		if ext != extJSON {
			return nil
		}

		if !utils.IsStringInSlice(ns, languages.Namespaces) {
			languages.Namespaces = append(languages.Namespaces, ns)
		}

		fdir, _ := filepath.Split(path)

		locale := filepath.Base(fdir)

		var localeReal string

		parts := strings.SplitN(locale, "-", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], parts[1]) {
			localeReal = parts[0]
		} else {
			localeReal = locale
		}

		if utils.IsStringInSlice(localeReal, locales) {
			for i, l := range languages.Languages {
				if l.Locale == localeReal {
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

		if tag, err = language.Parse(localeReal); err != nil {
			return fmt.Errorf("failed to parse language '%s': %w", localeReal, err)
		}

		l := Language{
			Display:    display.English.Tags().Name(tag),
			Locale:     localeReal,
			Namespaces: []string{ns},
			Fallbacks:  []string{languages.Defaults.Language.Locale},
			Tag:        tag,
		}

		languages.Languages = append(languages.Languages, l)

		locales = append(locales, l.Locale)

		return nil
	}); err != nil {
		return nil, err
	}

	var langs []Language //nolint:prealloc

	for i, lang := range languages.Languages {
		p := lang.Tag.Parent()

		if p.String() == "und" || strings.Contains(p.String(), "-") {
			continue
		}

		if utils.IsStringInSlice(p.String(), locales) {
			continue
		}

		if p.String() != lang.Locale {
			lang.Fallbacks = append([]string{p.String()}, lang.Fallbacks...)
		}

		languages.Languages[i] = lang

		l := Language{
			Display:    display.English.Tags().Name(p),
			Locale:     p.String(),
			Namespaces: lang.Namespaces,
			Fallbacks:  []string{languages.Defaults.Language.Locale},
			Tag:        p,
		}

		langs = append(langs, l)

		locales = append(locales, l.Locale)
	}

	languages.Languages = append(languages.Languages, langs...)

	sort.Slice(languages.Languages, func(i, j int) bool {
		return languages.Languages[i].Locale == localeDefault || languages.Languages[i].Locale < languages.Languages[j].Locale
	})

	return languages, nil
}
