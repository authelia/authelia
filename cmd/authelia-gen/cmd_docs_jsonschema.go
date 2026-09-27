// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/jsonschema"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

func newDocsJSONSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   pathJSONSchema,
		Short: "Generate docs JSON schema",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newDocsJSONSchemaConfigurationCmd(), newDocsJSONSchemaUserDatabaseCmd(), newDocsJSONSchemaExportsCmd())

	return cmd
}

func newDocsJSONSchemaExportsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exports",
		Short: "Generate docs JSON schema for the various exports",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newDocsJSONSchemaExportsTOTPCmd(), newDocsJSONSchemaExportsWebAuthnCmd(), newDocsJSONSchemaExportsIdentifiersCmd())

	return cmd
}

func newDocsJSONSchemaExportsTOTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "totp",
		Short: "Generate docs JSON schema for the TOTP exports",
		RunE:  docsJSONSchemaExportsTOTPRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newDocsJSONSchemaExportsWebAuthnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webauthn",
		Short: "Generate docs JSON schema for the WebAuthn exports",
		RunE:  docsJSONSchemaExportsWebAuthnRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newDocsJSONSchemaExportsIdentifiersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identifiers",
		Short: "Generate docs JSON schema for the identifiers exports",
		RunE:  docsJSONSchemaExportsIdentifiersRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newDocsJSONSchemaConfigurationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Generate docs JSON schema for the configuration",
		RunE:  docsJSONSchemaConfigurationRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newDocsJSONSchemaUserDatabaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-database",
		Short: "Generate docs JSON schema for the user database",
		RunE:  docsJSONSchemaUserDatabaseRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func docsJSONSchemaExportsTOTPRunE(cmd *cobra.Command, args []string) (err error) {
	var version *model.SemanticVersion

	if version, err = readVersion(cmd); err != nil {
		return err
	}

	var (
		dir, file, schemaDir string
	)

	if schemaDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirSchema); err != nil {
		return err
	}

	if dir, file, err = getJSONSchemaOutputPath(cmd, cmdFlagDocsStaticJSONSchemaExportsTOTP); err != nil {
		return err
	}

	return docsJSONSchemaGenerateRunE(cmd, args, version, schemaDir, &model.TOTPConfigurationDataExport{}, dir, file, nil)
}

func docsJSONSchemaExportsWebAuthnRunE(cmd *cobra.Command, args []string) (err error) {
	var version *model.SemanticVersion

	if version, err = readVersion(cmd); err != nil {
		return err
	}

	var (
		dir, file, schemaDir string
	)

	if schemaDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirSchema); err != nil {
		return err
	}

	if dir, file, err = getJSONSchemaOutputPath(cmd, cmdFlagDocsStaticJSONSchemaExportsWebAuthn); err != nil {
		return err
	}

	return docsJSONSchemaGenerateRunE(cmd, args, version, schemaDir, &model.WebAuthnCredentialDataExport{}, dir, file, nil)
}

func docsJSONSchemaExportsIdentifiersRunE(cmd *cobra.Command, args []string) (err error) {
	var version *model.SemanticVersion

	if version, err = readVersion(cmd); err != nil {
		return err
	}

	var (
		dir, file, schemaDir string
	)

	if schemaDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirSchema); err != nil {
		return err
	}

	if dir, file, err = getJSONSchemaOutputPath(cmd, cmdFlagDocsStaticJSONSchemaExportsIdentifiers); err != nil {
		return err
	}

	return docsJSONSchemaGenerateRunE(cmd, args, version, schemaDir, &model.UserOpaqueIdentifiersExport{}, dir, file, nil)
}

func docsJSONSchemaConfigurationRunE(cmd *cobra.Command, args []string) (err error) {
	var version *model.SemanticVersion

	if version, err = readVersion(cmd); err != nil {
		return err
	}

	var (
		dir, file, schemaDir string
	)

	if schemaDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirSchema); err != nil {
		return err
	}

	if dir, file, err = getJSONSchemaOutputPath(cmd, cmdFlagDocsStaticJSONSchemaConfiguration); err != nil {
		return err
	}

	return docsJSONSchemaGenerateRunE(cmd, args, version, schemaDir, &schema.Configuration{}, dir, file, jsonschemaKoanfMapper)
}

func docsJSONSchemaUserDatabaseRunE(cmd *cobra.Command, args []string) (err error) {
	var version *model.SemanticVersion

	if version, err = readVersion(cmd); err != nil {
		return err
	}

	var (
		dir, file, schemaDir string
	)

	if schemaDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirAuthentication); err != nil {
		return err
	}

	if dir, file, err = getJSONSchemaOutputPath(cmd, cmdFlagDocsStaticJSONSchemaUserDatabase); err != nil {
		return err
	}

	return docsJSONSchemaGenerateRunE(cmd, args, version, schemaDir, &authentication.FileUserDatabase{}, dir, file, jsonschemaKoanfMapper)
}

func docsJSONSchemaGenerateRunE(cmd *cobra.Command, _ []string, version *model.SemanticVersion, schemaDir string, v any, dir, file string, mapper func(reflect.Type) *jsonschema.Schema) (err error) {
	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		Mapper:                     mapper,
	}

	if runtime.GOOS == windows {
		mapComments := map[string]string{}

		if err = jsonschema.ExtractGoComments(goModuleBase, schemaDir, mapComments); err != nil {
			return err
		}

		if r.CommentMap == nil {
			r.CommentMap = map[string]string{}
		}

		for key, comment := range mapComments {
			r.CommentMap[strings.ReplaceAll(key, `\`, `/`)] = comment
		}
	} else {
		if err = r.AddGoComments(goModuleBase, schemaDir); err != nil {
			return err
		}
	}

	var (
		versions []string
		target   model.SemanticVersion
	)

	versions, _ = cmd.Flags().GetStringSlice(cmdFlagVersions)

	if len(versions) == 0 {
		versions = []string{metaVersionLatest, metaVersionCurrent}
	}

	if target, err = jsonSchemaVersionTarget(version, versions); err != nil {
		return err
	}

	schema := r.Reflect(v)

	jsonSchemaAddSchemaProperty(schema)

	for _, versionName := range versions {
		var out string

		switch versionName {
		case metaVersionMajor, metaVersionMinor:
			out = jsonSchemaVersionDir(target)
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, out, file))
		case metaVersionCurrent:
			out = jsonSchemaVersionDir(*version)
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, out, file))
		case metaVersionLatest:
			out = metaVersionLatest
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, jsonSchemaVersionDir(target), file))
		default:
			var parsed *model.SemanticVersion

			if parsed, err = model.NewSemanticVersion(versionName); err != nil {
				return fmt.Errorf("failed to parse version: %w", err)
			}

			out = jsonSchemaVersionDir(*parsed)
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, out, file))
		}

		if err = writeJSONSchema(schema, dir, out, file); err != nil {
			return err
		}
	}

	return nil
}

// jsonSchemaAddSchemaProperty adds the '$schema' property to the root object so JSON files may reference the schema
// without failing validation, as the root object does not allow additional properties.
func jsonSchemaAddSchemaProperty(schema *jsonschema.Schema) {
	root := schema

	if name, ok := strings.CutPrefix(schema.Ref, "#/$defs/"); ok {
		if def, ok := schema.Definitions[name]; ok {
			root = def
		}
	}

	if root.Properties == nil {
		return
	}

	root.Properties.Set("$schema", &jsonschema.Schema{
		Type:        "string",
		Format:      "uri",
		Title:       "JSON Schema",
		Description: "The JSON Schema which applies to this file.",
	})
}

func jsonSchemaVersionTarget(version *model.SemanticVersion, versions []string) (target model.SemanticVersion, err error) {
	major, minor := utils.IsStringInSlice(metaVersionMajor, versions), utils.IsStringInSlice(metaVersionMinor, versions)

	if major && minor {
		return target, fmt.Errorf("failed to generate: meta versions major and minor are mutually exclusive")
	}

	if (major || minor) && utils.IsStringInSlice(metaVersionCurrent, versions) {
		return target, fmt.Errorf("failed to generate: meta version current is mutually exclusive with major and minor")
	}

	switch {
	case major:
		return version.NextMajor(), nil
	case minor:
		return version.NextMinor(), nil
	default:
		return version.Copy(), nil
	}
}

func jsonSchemaVersionDir(version model.SemanticVersion) string {
	return fmt.Sprintf("v%d.%d", version.Major, version.Minor)
}

func writeJSONSchema(schema *jsonschema.Schema, dir, version, file string) (err error) {
	var (
		f *os.File
	)

	if _, err = os.Stat(filepath.Join(dir, version, pathJSONSchema)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Join(dir, version, pathJSONSchema), 0755); err != nil {
			return err
		}
	}

	if f, err = os.Create(filepath.Join(dir, version, pathJSONSchema, file+utils.ExtJSON)); err != nil {
		return err
	}

	encoder := json.NewEncoder(f)

	encoder.SetIndent("", "  ")

	if err = encoder.Encode(schema); err != nil {
		return err
	}

	return f.Close()
}

func getJSONSchemaOutputPath(cmd *cobra.Command, flag string) (dir, file string, err error) {
	if dir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsStatic, cmdFlagDocsStaticJSONSchemas); err != nil {
		return "", "", err
	}

	if file, err = cmd.Flags().GetString(flag); err != nil {
		return "", "", err
	}

	return dir, file, nil
}

func jsonschemaKoanfMapper(t reflect.Type) *jsonschema.Schema {
	switch t.String() {
	case "*language.Tag", "language.Tag":
		return &jsonschema.Schema{
			Type:    jsonschema.TypeString,
			Pattern: `^[a-z]{2}-[A-Z]{2}$`,
		}
	case "[]*net.IPNet":
		return &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{
					Type: jsonschema.TypeString,
				},
				{
					Type: jsonschema.TypeArray,
					Items: &jsonschema.Schema{
						Type: jsonschema.TypeString,
					},
				},
			},
		}
	case "regexp.Regexp", "*regexp.Regexp":
		return &jsonschema.Schema{
			Type:   jsonschema.TypeString,
			Format: jsonschema.FormatStringRegex,
		}
	case "time.Duration", "*time.Duration":
		return &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{
					Type:    jsonschema.TypeString,
					Pattern: `^\d+\s*(y|M|w|d|h|m|s|ms|((year|month|week|day|hour|minute|second|millisecond)s?))(\s*(\s+and\s+)?\d+\s*(y|M|w|d|h|m|s|ms|((year|month|week|day|hour|minute|second|millisecond)s?)))*$`,
				},
				{
					Type:        jsonschema.TypeInteger,
					Description: "The duration in seconds",
				},
			},
		}
	case "schema.CryptographicKey":
		return &jsonschema.Schema{
			Type:    jsonschema.TypeString,
			Pattern: `^-{5}BEGIN (((RSA|EC) )?(PRIVATE|PUBLIC) KEY|CERTIFICATE)-{5}\n([a-zA-Z0-9\/+]{1,64}\n)+([a-zA-Z0-9\/+]{1,64}[=]{0,2})\n-{5}END (((RSA|EC) )?(PRIVATE|PUBLIC) KEY|CERTIFICATE)-{5}\n?$`,
		}
	case "schema.CryptographicPrivateKey":
		return &jsonschema.Schema{
			Type:    jsonschema.TypeString,
			Pattern: `^-{5}BEGIN ((RSA|EC) )?PRIVATE KEY-{5}\n([a-zA-Z0-9\/+]{1,64}\n)+([a-zA-Z0-9\/+]{1,64}[=]{0,2})\n-{5}END ((RSA|EC) )?PRIVATE KEY-{5}\n?$`,
		}
	case "rsa.PrivateKey", "*rsa.PrivateKey":
		return &jsonschema.Schema{
			Type:    jsonschema.TypeString,
			Pattern: `^-{5}(BEGIN (RSA )?PRIVATE KEY-{5}\n([a-zA-Z0-9\/+]{1,64}\n)+([a-zA-Z0-9\/+]{1,64}[=]{0,2})\n-{5}END (RSA )?PRIVATE KEY-{5}\n?)+$`,
		}
	case "ecdsa.PrivateKey", "*.ecdsa.PrivateKey":
		return &jsonschema.Schema{
			Type:    jsonschema.TypeString,
			Pattern: `^-{5}(BEGIN ((EC )?PRIVATE KEY-{5}\n([a-zA-Z0-9\/+]{1,64}\n)+([a-zA-Z0-9\/+]{1,64}[=]{0,2})\n-{5}END (EC )?PRIVATE KEY-{5}\n?)+$`,
		}
	case "mail.Address", "*mail.Address":
		return &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{
					Type:   jsonschema.TypeString,
					Format: jsonschema.FormatStringEmail,
				},
				{
					Type:    jsonschema.TypeString,
					Pattern: `^[^<]+\s\<[a-zA-Z0-9._~!#$%&'*/=?^{|}+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9-]+\>$`,
				},
			},
		}
	case "schema.CSPTemplate":
		return &jsonschema.Schema{
			Type:    jsonschema.TypeString,
			Default: buildCSP(codeCSPSelf, codeCSPValuesCommon, codeCSPValuesProduction),
		}
	}

	return nil
}
