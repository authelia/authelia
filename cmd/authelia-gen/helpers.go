package main

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/mail"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
	reflect.TypeOf(schema.X509CertificateChain{}),
	reflect.TypeOf(schema.PasswordDigest{}),
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

//nolint:gocyclo
func readTags(prefix string, t reflect.Type, envSkip bool) (tags []string) {
	tags = make([]string, 0)

	if envSkip && (t.Kind() == reflect.Slice || t.Kind() == reflect.Map) {
		return
	}

	if t.Kind() != reflect.Struct {
		if t.Kind() == reflect.Slice {
			tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, "", true, false), t.Elem(), envSkip)...)
		}

		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag := field.Tag.Get("koanf")

		if tag == "" {
			tags = append(tags, prefix)

			continue
		}

		switch kind := field.Type.Kind(); kind {
		case reflect.Struct:
			if !containsType(field.Type, decodedTypes) {
				tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, false, false), field.Type, envSkip)...)

				continue
			}
		case reflect.Slice, reflect.Map:
			if envSkip {
				continue
			}

			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				if !containsType(field.Type.Elem(), decodedTypes) {
					tags = append(tags, getKeyNameFromTagAndPrefix(prefix, tag, false, false))
					tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, kind == reflect.Slice, kind == reflect.Map), field.Type.Elem(), envSkip)...)

					continue
				}
			case reflect.Slice:
				tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, kind == reflect.Slice, kind == reflect.Map), field.Type.Elem(), envSkip)...)
			}
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				if !containsType(field.Type.Elem(), decodedTypes) {
					tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, false, false), field.Type.Elem(), envSkip)...)

					continue
				}
			case reflect.Slice, reflect.Map:
				if envSkip {
					continue
				}

				if field.Type.Elem().Elem().Kind() == reflect.Struct {
					if !containsType(field.Type.Elem(), decodedTypes) {
						tags = append(tags, readTags(getKeyNameFromTagAndPrefix(prefix, tag, true, false), field.Type.Elem(), envSkip)...)

						continue
					}
				}
			}
		}

		tags = append(tags, getKeyNameFromTagAndPrefix(prefix, tag, false, false))
	}

	return tags
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
