// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/jsonschema"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
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

	cmd.AddCommand(newDocsJSONSchemaConfigurationCmd(), newDocsJSONSchemaUserDatabaseCmd(), newDocsJSONSchemaExportsCmd(), newDocsJSONSchemaWebhooksCmd())

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

func newDocsJSONSchemaWebhooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   dirDocsStaticJSONSchemasWebhooks,
		Short: "Generate docs JSON schema for the webhook event payloads",
		RunE:  docsJSONSchemaWebhooksRunE,

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

func docsJSONSchemaWebhooksRunE(cmd *cobra.Command, _ []string) (err error) {
	var dir, eventsDir string

	if eventsDir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDirEvents); err != nil {
		return err
	}

	if dir, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsStatic, cmdFlagDocsStaticJSONSchemas, cmdFlagDocsStaticJSONSchemasWebhooks); err != nil {
		return err
	}

	var r *jsonschema.Reflector

	if r, err = newWebhookJSONSchemaReflector(eventsDir); err != nil {
		return err
	}

	dir = webhookJSONSchemaDir(dir)

	for _, document := range webhookJSONSchemaDocuments() {
		if err = writeJSONSchemaFile(reflectWebhookJSONSchema(r, document.Value, document.Name), dir, document.Name+extJSON); err != nil {
			return err
		}
	}

	return nil
}

type webhookJSONSchemaDocument struct {
	Name  string
	Value any
}

func webhookJSONSchemaDocuments() (documents []webhookJSONSchemaDocument) {
	names := events.Registered()

	documents = make([]webhookJSONSchemaDocument, 0, len(names)+2)
	documents = append(documents,
		webhookJSONSchemaDocument{Name: fileDocsStaticJSONSchemasWebhooksEnvelope, Value: &events.Envelope{}},
		webhookJSONSchemaDocument{Name: fileDocsStaticJSONSchemasWebhooksBatch, Value: &events.EnvelopeBatch{}},
	)

	for _, name := range names {
		descriptor, ok := events.Lookup(name)
		if !ok {
			continue
		}

		documents = append(documents, webhookJSONSchemaDocument{Name: name, Value: descriptor.New()})
	}

	return documents
}

func webhookJSONSchemaDir(dir string) string {
	return filepath.Join(dir, "v"+events.ContractVersion)
}

func newWebhookJSONSchemaReflector(dir string) (r *jsonschema.Reflector, err error) {
	r = &jsonschema.Reflector{
		Anonymous:                 true,
		ExpandedStruct:            true,
		AllowAdditionalProperties: true,
	}

	if err = jsonschemaAddGoComments(r, dir); err != nil {
		return nil, err
	}

	return r, nil
}

func reflectWebhookJSONSchema(r *jsonschema.Reflector, v any, name string) (s *jsonschema.Schema) {
	s = r.Reflect(v)

	s.ID = jsonschema.ID(fmt.Sprintf(urlFormatJSONSchemaWebhooks, events.ContractVersion, name))

	return s
}

func docsJSONSchemaGenerateRunE(cmd *cobra.Command, _ []string, version *model.SemanticVersion, schemaDir string, v any, dir, file string, mapper func(reflect.Type) *jsonschema.Schema) (err error) {
	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		Mapper:                     mapper,
	}

	if err = jsonschemaAddGoComments(r, schemaDir); err != nil {
		return err
	}

	var (
		versions []string
	)

	versions, _ = cmd.Flags().GetStringSlice(cmdFlagVersions)

	if len(versions) == 0 {
		versions = []string{metaVersionLatest, metaVersionCurrent}
	}

	next := utils.IsStringInSlice(metaVersionNext, versions)

	if next && utils.IsStringInSlice(metaVersionCurrent, versions) {
		return fmt.Errorf("failed to generate: meta version next and current are mutually exclusive")
	}

	schema := r.Reflect(v)

	for _, versionName := range versions {
		var out string

		switch versionName {
		case metaVersionNext:
			out = fmt.Sprintf("v%d.%d", version.Major, version.Minor+1)
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, out, file))
		case metaVersionCurrent:
			out = fmt.Sprintf("v%d.%d", version.Major, version.Minor)
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, out, file))
		case metaVersionLatest:
			out = metaVersionLatest

			if next {
				schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, fmt.Sprintf("v%d.%d", version.Major, version.Minor+1), file))
			} else {
				schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, fmt.Sprintf("v%d.%d", version.Major, version.Minor), file))
			}
		default:
			var parsed *model.SemanticVersion

			if parsed, err = model.NewSemanticVersion(versionName); err != nil {
				return fmt.Errorf("failed to parse version: %w", err)
			}

			out = fmt.Sprintf("v%d.%d", parsed.Major, parsed.Minor)
			schema.ID = jsonschema.ID(fmt.Sprintf(model.FormatJSONSchemaIdentifier, fmt.Sprintf("v%d.%d", version.Major, version.Minor), file))
		}

		if err = writeJSONSchema(schema, dir, out, file); err != nil {
			return err
		}
	}

	return nil
}

func jsonschemaAddGoComments(r *jsonschema.Reflector, dir string) (err error) {
	if runtime.GOOS != windows {
		return r.AddGoComments(goModuleBase, dir)
	}

	mapComments := map[string]string{}

	if err = jsonschema.ExtractGoComments(goModuleBase, dir, mapComments); err != nil {
		return err
	}

	if r.CommentMap == nil {
		r.CommentMap = map[string]string{}
	}

	for key, comment := range mapComments {
		r.CommentMap[strings.ReplaceAll(key, `\`, `/`)] = comment
	}

	return nil
}

func writeJSONSchema(schema *jsonschema.Schema, dir, version, file string) (err error) {
	return writeJSONSchemaFile(schema, filepath.Join(dir, version, pathJSONSchema), file+extJSON)
}

func writeJSONSchemaFile(schema *jsonschema.Schema, dir, file string) (err error) {
	var (
		f *os.File
	)

	if _, err = os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if f, err = os.Create(filepath.Join(dir, file)); err != nil {
		return err
	}

	if err = encodeJSONSchema(f, schema); err != nil {
		return err
	}

	return f.Close()
}

func encodeJSONSchema(w io.Writer, schema *jsonschema.Schema) (err error) {
	encoder := json.NewEncoder(w)

	encoder.SetIndent("", "  ")

	return encoder.Encode(schema)
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
