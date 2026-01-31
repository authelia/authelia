package utils

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// GetDirectoryLanguages returns the languages information from a given directory.
func GetDirectoryLanguages(path string) (languages *Languages, err error) {
	fileSystem := os.DirFS(path)

	lng, err := getLanguages(fileSystem)
	if err != nil {
		return nil, err
	}

	return lng, nil
}

// GetEmbeddedLanguages returns the available languages info from an embed.FS.
func GetEmbeddedLanguages(fs embed.FS) (languages *Languages, err error) {
	return getLanguages(fs)
}

//nolint:gocyclo
func getLanguages(dir fs.FS) (languages *Languages, err error) {
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

	caser := cases.Title(defaultTag)

	languages.Defaults.Language = Language{
		Display: caser.String(display.Self.Name(defaultTag)),
		Locale:  localeDefault,
	}

	if err = fs.WalkDir(dir, ".", func(path string, info fs.DirEntry, errWalk error) (err error) {
		if errWalk != nil {
			return errWalk
		}
		if info.IsDir() {
			return nil
		}

		nameLower := strings.ToLower(info.Name())
		ext := filepath.Ext(nameLower)

		ns := strings.Replace(nameLower, ext, "", 1)

		if ext != extJSON {
			return nil
		}

		if !IsStringInSlice(ns, languages.Namespaces) {
			languages.Namespaces = append(languages.Namespaces, ns)
		}

		fdir, _ := filepath.Split(path)

		locale := filepath.Base(fdir)

		if IsStringInSlice(locale, locales) {
			for i, l := range languages.Languages {
				if l.Locale == locale {
					if IsStringInSlice(ns, languages.Languages[i].Namespaces) {
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
			return fmt.Errorf("failed to parse language '%s': %w", locale, err)
		}

		langDisplay := cases.Title(tag)

		l := Language{
			Display:    langDisplay.String(display.Self.Name(tag)),
			Locale:     locale,
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

	var langs []Language

	for i, lang := range languages.Languages {
		p := lang.Tag.Parent()
		parentString := p.String()

		if parentString == undefinedLocaleTag {
			continue
		}

		// if parent is a microlanguage, use base language as parent
		// this is true for cases as spanish argentina and spanish mexico.
		if strings.Contains(parentString, "-") {
			b, _ := lang.Tag.Base()
			if p, err = language.Parse(b.String()); err != nil {
				continue
			}

			parentString = p.String()
		}

		if parentString != lang.Locale {
			lang.Fallbacks = append([]string{parentString}, lang.Fallbacks...)
			lang.Parent = parentString
		}

		languages.Languages[i] = lang

		if IsStringInSlice(parentString, locales) {
			continue
		}

		fallbackDisplay := cases.Title(lang.Tag)
		l := Language{
			Display:    fallbackDisplay.String(display.Self.Name(p)),
			Locale:     parentString,
			Namespaces: lang.Namespaces,
			Fallbacks:  []string{languages.Defaults.Language.Locale},
			Tag:        p,
		}

		langs = append(langs, l)

		locales = append(locales, l.Locale)
	}

	languages.Languages = append(languages.Languages, langs...)

	sort.Slice(languages.Languages, func(i, j int) bool {
		if languages.Languages[i].Locale == localeDefault {
			return true
		}

		if languages.Languages[j].Locale == localeDefault {
			return false
		}

		return languages.Languages[i].Locale < languages.Languages[j].Locale
	})

	return languages, nil
}

// GetLocaleParentOrBaseString returns the base language specified locale.
func GetLocaleParentOrBaseString(locale string) (string, error) {
	tag, err := language.Parse(locale)
	if err != nil {
		return "", fmt.Errorf("failed to parse language '%s': %w", locale, err)
	}

	parent := tag.Parent()

	if parent.String() == undefinedLocaleTag || strings.Contains(parent.String(), "-") {
		base, _ := tag.Base()

		return base.String(), nil
	}

	return parent.String(), nil
}
