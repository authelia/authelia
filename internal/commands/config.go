package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newConfigCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "config",
		Short:   cmdAutheliaConfigShort,
		Long:    cmdAutheliaConfigLong,
		Example: cmdAutheliaConfigExample,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
		),
		RunE: ctx.ConfigValidateRunE,

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newConfigValidateCmd(ctx), newConfigTemplateCmd(ctx))

	return cmd
}

func newConfigTemplateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "template",
		Short:   cmdAutheliaConfigTemplateShort,
		Long:    cmdAutheliaConfigTemplateLong,
		Example: cmdAutheliaConfigTemplateExample,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
		),
		RunE: ctx.ConfigTemplateRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func newConfigValidateCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "validate",
		Short:   cmdAutheliaConfigValidateShort,
		Long:    cmdAutheliaConfigValidateLong,
		Example: cmdAutheliaConfigValidateExample,
		Args:    cobra.NoArgs,
		PreRunE: ctx.ChainRunE(
			ctx.HelperConfigLoadRunE,
			ctx.HelperConfigValidateKeysRunE,
			ctx.HelperConfigValidateRunE,
		),
		RunE: ctx.ConfigValidateRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

// ConfigValidateRunE is the RunE for the authelia validate-config command.
func (ctx *CmdCtx) ConfigValidateRunE(cmd *cobra.Command, _ []string) (err error) {
	return runConfigValidate(cmd.OutOrStdout(), ctx.cconfig.validator)
}

func runConfigValidate(w io.Writer, validator *schema.StructValidator) (err error) {
	var isError bool

	switch {
	case validator.HasErrors():
		isError = true

		_, _ = fmt.Fprintf(w, "Configuration parsed and loaded with errors:\n\n")

		for _, err = range validator.Errors() {
			_, _ = fmt.Fprintf(w, "\t - %v\n", err)
		}

		_, _ = fmt.Fprint(w, "\n")

		if !validator.HasWarnings() {
			break
		}

		fallthrough
	case validator.HasWarnings():
		_, _ = fmt.Fprintf(w, "Configuration parsed and loaded with warnings:\n\n")

		for _, err = range validator.Warnings() {
			_, _ = fmt.Fprintf(w, "\t - %v\n", err)
		}

		_, _ = fmt.Fprint(w, "\n")
	default:
		_, _ = fmt.Fprintf(w, "Configuration parsed and loaded successfully without errors.\n\n")
	}

	if isError {
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

// ConfigTemplateRunE is the RunE for the authelia validate-config command.
func (ctx *CmdCtx) ConfigTemplateRunE(cmd *cobra.Command, _ []string) (err error) {
	return runConfigTemplate(cmd.OutOrStdout(), ctx.cconfig.sources)
}

func runConfigTemplate(w io.Writer, sources []configuration.Source) (err error) {
	var (
		source *configuration.FileSource
		ok     bool
	)

	var files []*configuration.File

	var n int

	for _, s := range sources {
		if source, ok = s.(*configuration.FileSource); !ok {
			continue
		}

		n++

		if files, err = source.ReadFiles(); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, fmtYAMLConfigTemplateHeader, strings.Join(source.GetBytesFilterNames(), ", "))

		for _, file := range files {
			if reYAMLComment.Match(file.Data) {
				_, _ = w.Write(reYAMLComment.ReplaceAll(file.Data, []byte(fmt.Sprintf(fmtYAMLConfigTemplateFileHeader+"$1", file.Path))))
			} else {
				_, _ = fmt.Fprintf(w, fmtYAMLConfigTemplateFileHeader, file.Path)
				_, _ = w.Write(file.Data)
			}
		}
	}

	if n == 0 {
		return fmt.Errorf("templating requires configuration files however no configuration file sources were specified")
	}

	return nil
}

func newConfigValidateLegacyCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = newConfigValidateCmd(ctx)

	cmd.Use = "validate-config"
	cmd.Example = cmdAutheliaConfigValidateLegacyExample
	cmd.Hidden = true

	return cmd
}
