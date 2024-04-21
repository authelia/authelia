package main

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
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
	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

func getPFlagPath(flags *pflag.FlagSet, flagNames ...string) (fullPath string, err error) {
	if len(flagNames) == 0 {
		return "", fmt.Errorf("no flag names")
	}

	var p string

	for i, flagName := range flagNames {
		if p, err = flags.GetString(flagName); err != nil {
			return "", fmt.Errorf("failed to lookup flag '%s': %w", flagName, err)
		}

		if i == 0 {
			fullPath = p
		} else {
			fullPath = filepath.Join(fullPath, p)
		}
	}

	return fullPath, nil
}

func buildCSP(defaultSrc string, ruleSets ...[]CSPValue) string {
	var rules []string

	for _, ruleSet := range ruleSets {
		for _, rule := range ruleSet {
			switch rule.Name {
			case "default-src":
				rules = append(rules, fmt.Sprintf("%s %s", rule.Name, defaultSrc))
			default:
				rules = append(rules, fmt.Sprintf("%s %s", rule.Name, rule.Value))
			}
		}
	}

	return strings.Join(rules, "; ")
}

var decodedTypes = []reflect.Type{
	reflect.TypeOf(mail.Address{}),
	reflect.TypeOf(regexp.Regexp{}),
	reflect.TypeOf(url.URL{}),
	reflect.TypeOf(time.Duration(0)),
	reflect.TypeOf(schema.Address{}),
	reflect.TypeOf(schema.AddressTCP{}),
	reflect.TypeOf(schema.AddressUDP{}),
	reflect.TypeOf(schema.AddressLDAP{}),
	reflect.TypeOf(schema.AddressSMTP{}),
	reflect.TypeOf(schema.X509CertificateChain{}),
	reflect.TypeOf(schema.PasswordDigest{}),
	reflect.TypeOf(schema.RefreshIntervalDuration{}),
	reflect.TypeOf(rsa.PrivateKey{}),
	reflect.TypeOf(ecdsa.PrivateKey{}),
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

func readVersion(cmd *cobra.Command) (version *model.SemanticVersion, err error) {
	var (
		pathPackageJSON string
		dataPackageJSON []byte
		packageJSON     PackageJSON
	)

	if pathPackageJSON, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagWeb, cmdFlagFileWebPackage); err != nil {
		return nil, err
	}

	if dataPackageJSON, err = os.ReadFile(pathPackageJSON); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(dataPackageJSON, &packageJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshall package.json: %w", err)
	}

	return model.NewSemanticVersion(packageJSON.Version)
}

//nolint:gocyclo
func readTags(prefix string, t reflect.Type, envSkip, deprecatedSkip bool) (tags []string) {
	tags = make([]string, 0)

	if envSkip && (t.Kind() == reflect.Slice || t.Kind() == reflect.Map) {
		return
	}

	if t.Kind() != reflect.Struct {
		if t.Kind() == reflect.Slice {
			tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, "", true, false), t.Elem(), envSkip, deprecatedSkip)...)
		}

		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if deprecatedSkip && isDeprecated(field) {
			continue
		}

		tag := field.Tag.Get("koanf")

		if tag == "" {
			tags = append(tags, prefix)

			continue
		}

		switch kind := field.Type.Kind(); kind {
		case reflect.Struct:
			if !containsType(field.Type, decodedTypes) {
				tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, false, false), field.Type, envSkip, deprecatedSkip)...)

				continue
			}
		case reflect.Slice, reflect.Map:
			k := field.Type.Elem().Kind()

			if envSkip && !isValueKind(k) {
				continue
			}

			switch k {
			case reflect.Struct:
				if !containsType(field.Type.Elem(), decodedTypes) {
					tags = append(tags, getKeyNameFromTagAndPrefix(prefix, tag, false, false))
					tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, kind == reflect.Slice, kind == reflect.Map), field.Type.Elem(), envSkip, deprecatedSkip)...)

					continue
				}
			case reflect.Slice:
				tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, kind == reflect.Slice, kind == reflect.Map), field.Type.Elem(), envSkip, deprecatedSkip)...)
			}
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				if !containsType(field.Type.Elem(), decodedTypes) {
					tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, false, false), field.Type.Elem(), envSkip, deprecatedSkip)...)

					continue
				}
			case reflect.Slice, reflect.Map:
				k := field.Type.Elem().Elem().Kind()

				if envSkip && !isValueKind(k) {
					continue
				}

				if k == reflect.Struct {
					if !containsType(field.Type.Elem(), decodedTypes) {
						tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, true, false), field.Type.Elem(), envSkip, deprecatedSkip)...)

						continue
					}
				}
			}
		}

		tags = append(tags, getKeyNameFromTagAndPrefix(prefix, tag, false, false))
	}

	return tags
}

func isValueKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Pointer, reflect.UnsafePointer, reflect.Invalid, reflect.Uintptr:
		return false
	default:
		return true
	}
}

func isDeprecated(field reflect.StructField) bool {
	var (
		value string
		ok    bool
	)

	if value, ok = field.Tag.Lookup("jsonschema"); !ok {
		return false
	}

	return utils.IsStringInSlice("deprecated", strings.Split(value, ","))
}

func getKeyNameFromTagAndPrefix(prefix, name string, isSlice, isMap bool) string {
	nameParts := strings.SplitN(name, ",", 2)

	if prefix == "" {
		return nameParts[0]
	}

	if len(nameParts) == 2 && nameParts[1] == "squash" {
		return prefix
	}

	switch {
	case isMap:
		if name == "" {
			return fmt.Sprintf("%s.*", prefix)
		}

		return fmt.Sprintf("%s.%s.*", prefix, nameParts[0])
	case isSlice:
		if name == "" {
			return fmt.Sprintf("%s[]", prefix)
		}

		return fmt.Sprintf("%s.%s[]", prefix, nameParts[0])
	default:
		return fmt.Sprintf("%s.%s", prefix, nameParts[0])
	}
}
