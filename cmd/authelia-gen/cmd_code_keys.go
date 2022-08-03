package main

import (
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newCodeKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Generate the list of valid configuration keys",
		RunE:  codeKeysRunE,
	}

	cmd.Flags().StringP("file", "f", "./internal/configuration/schema/keys.go", "Sets the path of the keys file")
	cmd.Flags().String("package", "schema", "Sets the package name of the keys file")

	return cmd
}

func codeKeysRunE(cmd *cobra.Command, _ []string) (err error) {
	var (
		file string

		f *os.File
	)

	data := keysTemplateStruct{
		Timestamp: time.Now(),
		Keys:      readTags("", reflect.TypeOf(schema.Configuration{})),
	}

	if file, err = cmd.Flags().GetString("file"); err != nil {
		return err
	}

	if data.Package, err = cmd.Flags().GetString("package"); err != nil {
		return err
	}

	if f, err = os.Create(file); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", file, err)
	}

	var (
		content []byte
		tmpl    *template.Template
	)

	if content, err = templatesFS.ReadFile("templates/config_keys.go.tmpl"); err != nil {
		return err
	}

	if tmpl, err = template.New("keys").Parse(string(content)); err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

type keysTemplateStruct struct {
	Timestamp time.Time
	Keys      []string
	Package   string
}

var decodedTypes = []reflect.Type{
	reflect.TypeOf(mail.Address{}),
	reflect.TypeOf(regexp.Regexp{}),
	reflect.TypeOf(url.URL{}),
	reflect.TypeOf(time.Duration(0)),
	reflect.TypeOf(schema.Address{}),
}

func containsType(needle reflect.Type, haystack []reflect.Type) (contains bool) {
	for _, t := range haystack {
		if needle.Kind() == reflect.Ptr {
			if needle.Elem() == t {
				return true
			}
		} else if needle == t {
			return true
		}
	}

	return false
}

func readTags(prefix string, t reflect.Type) (tags []string) {
	tags = make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag := field.Tag.Get("koanf")

		if tag == "" {
			tags = append(tags, prefix)

			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if !containsType(field.Type, decodedTypes) {
				tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, false), field.Type)...)

				continue
			}
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.Struct {
				if !containsType(field.Type.Elem(), decodedTypes) {
					tags = append(tags, getKeyNameFromTagAndPrefix(prefix, tag, false))
					tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, true), field.Type.Elem())...)

					continue
				}
			}
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				if !containsType(field.Type.Elem(), decodedTypes) {
					tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, false), field.Type.Elem())...)

					continue
				}
			case reflect.Slice:
				if field.Type.Elem().Elem().Kind() == reflect.Struct {
					if !containsType(field.Type.Elem(), decodedTypes) {
						tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, true), field.Type.Elem())...)

						continue
					}
				}
			}
		}

		tags = append(tags, getKeyNameFromTagAndPrefix(prefix, tag, false))
	}

	return tags
}

func getKeyNameFromTagAndPrefix(prefix, name string, slice bool) string {
	nameParts := strings.SplitN(name, ",", 2)

	if prefix == "" {
		return nameParts[0]
	}

	if len(nameParts) == 2 && nameParts[1] == "squash" {
		return prefix
	}

	if slice {
		return fmt.Sprintf("%s.%s[]", prefix, nameParts[0])
	}

	return fmt.Sprintf("%s.%s", prefix, nameParts[0])
}
