package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		Use:               "code",
		Short:             "Generate code",
		RunE:              rootSubCommandsRunE,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newCodeKeysCmd(), newCodeScriptsCmd())

	return cmd
}

func newCodeScriptsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "scripts",
		Short:             "Generate the generated portion of the authelia-scripts command",
		RunE:              codeScriptsRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func newCodeKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "keys",
		Short:             "Generate the list of valid configuration keys",
		RunE:              codeKeysRunE,
		DisableAutoGenTag: true,
	}

	return cmd
}

func codeScriptsRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		root, pathScriptsGen string
		resp                 *http.Response
	)

	data := &tmplScriptsGEnData{}

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathScriptsGen, err = cmd.Flags().GetString(cmdFlagFileScriptsGen); err != nil {
		return err
	}

	if data.Package, err = cmd.Flags().GetString(cmdFlagPackageScriptsGen); err != nil {
		return err
	}

	if resp, err = http.Get("https://api.github.com/repos/swagger-api/swagger-ui/tags"); err != nil {
		return fmt.Errorf("failed to get latest version of the Swagger UI: %w", err)
	}

	defer resp.Body.Close()

	var (
		respJSON []GitHubTagsJSON
		respRaw  []byte
	)

	if respRaw, err = io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("failed to get latest version of the Swagger UI: %w", err)
	}

	if err = json.Unmarshal(respRaw, &respJSON); err != nil {
		return fmt.Errorf("failed to get latest version of the Swagger UI: %w", err)
	}

	if len(respJSON) < 1 {
		return fmt.Errorf("failed to get latest version of the Swagger UI: the api returned zero results")
	}

	data.VersionSwaggerUI = respJSON[0].Name

	fullPathScriptsGen := filepath.Join(root, pathScriptsGen)

	var f *os.File

	if f, err = os.Create(fullPathScriptsGen); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", fullPathScriptsGen, err)
	}

	if err = tmplScriptsGen.Execute(f, data); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to write output file '%s': %w", fullPathScriptsGen, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close output file '%s': %w", fullPathScriptsGen, err)
	}

	return nil
}

type GitHubTagsJSON struct {
	Name string `json:"name"`
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
