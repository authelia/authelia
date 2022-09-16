package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDocsKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseKeys,
		Short: "Generate the list of valid configuration keys",
		RunE:  docsKeysRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func docsKeysRunE(cmd *cobra.Command, args []string) (err error) {
	//nolint:prealloc
	var (
		pathDocsConfigKeys, root string
		data                     []ConfigurationKey
	)

	keys := readTags("", reflect.TypeOf(schema.Configuration{}))

	for _, key := range keys {
		if strings.Contains(key, "[]") {
			continue
		}

		ck := ConfigurationKey{
			Path:   key,
			Secret: configuration.IsSecretKey(key),
		}

		switch {
		case ck.Secret:
			ck.Env = configuration.ToEnvironmentSecretKey(key, configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter)
		default:
			ck.Env = configuration.ToEnvironmentKey(key, configuration.DefaultEnvPrefix, configuration.DefaultEnvDelimiter)
		}

		data = append(data, ck)
	}

	var (
		dataJSON []byte
	)

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
		return err
	}

	if pathDocsConfigKeys, err = cmd.Flags().GetString(cmdFlagFileDocsKeys); err != nil {
		return err
	}

	fullPathDocsConfigKeys := filepath.Join(root, pathDocsConfigKeys)

	if dataJSON, err = json.Marshal(data); err != nil {
		return err
	}

	if err = os.WriteFile(fullPathDocsConfigKeys, dataJSON, 0600); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", fullPathDocsConfigKeys, err)
	}

	return nil
}
