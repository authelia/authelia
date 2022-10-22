package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
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
