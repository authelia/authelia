package main

import (
	"fmt"
	"path/filepath"

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
