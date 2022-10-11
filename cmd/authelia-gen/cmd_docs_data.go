package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDocsDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocsData,
		Short: "Generate docs data files",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newDocsDataMiscCmd(), newDocsDataKeysCmd())

	return cmd
}

func newDocsDataMiscCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseDocsDataMisc,
		Short: "Generate docs data file misc.json",
		RunE:  docsDataMiscRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func docsDataMiscRunE(cmd *cobra.Command, args []string) (err error) {
	data := DocsDataMisc{
		CSP: TemplateCSP{
			PlaceholderNONCE:    codeCSPNonce,
			TemplateDefault:     fmt.Sprintf(codeTmplCSPDefault, "", codeCSPNonce),
			TemplateDevelopment: fmt.Sprintf(codeTmplCSPDefault, " 'unsafe-eval'", codeCSPNonce),
		},
	}

	var (
		outputPath string
		dataJSON   []byte
	)

	if outputPath, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsData, cmdFlagDocsDataMisc); err != nil {
		return err
	}

	if dataJSON, err = json.Marshal(data); err != nil {
		return err
	}

	if err = os.WriteFile(outputPath, dataJSON, 0600); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", outputPath, err)
	}

	return nil
}

func newDocsDataKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseKeys,
		Short: "Generate the docs data file for configuration keys",
		RunE:  docsKeysRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func docsKeysRunE(cmd *cobra.Command, args []string) (err error) {
	//nolint:prealloc
	var (
		data []ConfigurationKey
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
		dataJSON   []byte
		outputPath string
	)

	if outputPath, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsData, cmdFlagDocsDataKeys); err != nil {
		return err
	}

	if dataJSON, err = json.Marshal(data); err != nil {
		return err
	}

	if err = os.WriteFile(outputPath, dataJSON, 0600); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", outputPath, err)
	}

	return nil
}
