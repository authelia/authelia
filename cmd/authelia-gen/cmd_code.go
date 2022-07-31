package main

import (
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code",
		Short: "Generate code",
		RunE:  rootSubCommandsRunE,
	}

	cmd.AddCommand(newCodeKeysCmd())

	cmd.PersistentFlags().String(cmdFlagFileConfigKeys, fileCodeConfigKeys, "Sets the path of the keys file")
	cmd.PersistentFlags().String(cmdFlagPackageConfigKeys, "schema", "Sets the package name of the keys file")

	return cmd
}

func newCodeKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Generate the list of valid configuration keys",
		RunE:  codeKeysRunE,
	}

	return cmd
}

func codeKeysRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		pathCodeConfigKeys, root string

		f *os.File
	)

	data := tmplConfigurationKeysData{
		Timestamp: time.Now(),
		Keys:      readTags("", reflect.TypeOf(schema.Configuration{})),
	}

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathCodeConfigKeys, err = cmd.Flags().GetString(cmdFlagFileConfigKeys); err != nil {
		return err
	}

	if data.Package, err = cmd.Flags().GetString(cmdFlagPackageConfigKeys); err != nil {
		return err
	}

	fullPathCodeConfigKeys := filepath.Join(root, pathCodeConfigKeys)

	if f, err = os.Create(fullPathCodeConfigKeys); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPathCodeConfigKeys, err)
	}

	if err = tmplCodeConfigurationSchemaKeys.Execute(f, data); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to write output file '%s': %w", fullPathCodeConfigKeys, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close output file '%s': %w", fullPathCodeConfigKeys, err)
	}

	return nil
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
