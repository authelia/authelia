package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseCode,
		Short: "Generate code",
		RunE:  rootSubCommandsRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newCodeKeysCmd(), newCodeServerCmd(), newCodeScriptsCmd())

	return cmd
}

func newCodeServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseServer,
		Short: "Generate the Authelia server files",
		RunE:  codeServerRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newCodeScriptsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseCodeScripts,
		Short: "Generate the generated portion of the authelia-scripts command",
		RunE:  codeScriptsRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newCodeKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseKeys,
		Short: "Generate the list of valid configuration keys",
		RunE:  codeKeysRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func codeServerRunE(cmd *cobra.Command, args []string) (err error) {
	data := TemplateCSP{
		PlaceholderNONCE:    codeCSPNonce,
		TemplateDefault:     buildCSP(codeCSPProductionDefaultSrc, codeCSPValuesCommon, codeCSPValuesProduction),
		TemplateDevelopment: buildCSP(codeCSPDevelopmentDefaultSrc, codeCSPValuesCommon, codeCSPValuesDevelopment),
	}

	var outputPath string

	if outputPath, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagFileServerGenerated); err != nil {
		return err
	}

	var f *os.File

	if f, err = os.Create(outputPath); err != nil {
		return fmt.Errorf("failed to create file '%s': %w", outputPath, err)
	}

	if err = tmplServer.Execute(f, data); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to write output file '%s': %w", outputPath, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close output file '%s': %w", outputPath, err)
	}

	return nil
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

	if resp, err = http.Get("https://api.github.com/repos/swagger-api/swagger-ui/releases/latest"); err != nil {
		return fmt.Errorf("failed to get latest version of the Swagger UI: %w", err)
	}

	defer resp.Body.Close()

	var (
		respJSON GitHubReleasesJSON
		respRaw  []byte
	)

	if respRaw, err = io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("failed to get latest version of the Swagger UI: %w", err)
	}

	if err = json.Unmarshal(respRaw, &respJSON); err != nil {
		return fmt.Errorf("failed to get latest version of the Swagger UI: %w", err)
	}

	if strings.HasPrefix(respJSON.TagName, "v") {
		data.VersionSwaggerUI = respJSON.TagName[1:]
	} else {
		data.VersionSwaggerUI = respJSON.TagName
	}

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

func codeKeysRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		pathCodeConfigKeys, root string

		f *os.File
	)

	data := tmplConfigurationKeysData{
		Timestamp: time.Now(),
		Keys:      readTags("", reflect.TypeOf(schema.Configuration{}), false, false, true),
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
