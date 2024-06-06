package commands

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/configuration"
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
func (ctx *CmdCtx) ConfigValidateRunE(_ *cobra.Command, _ []string) (err error) {
	var isError bool

	buf := &bytes.Buffer{}

	switch {
	case ctx.cconfig.validator.HasErrors():
		isError = true

		_, _ = fmt.Fprintf(buf, "Configuration parsed and loaded with errors:\n\n")

		for _, err = range ctx.cconfig.validator.Errors() {
			_, _ = fmt.Fprintf(buf, "\t - %v\n", err)
		}

		_, _ = fmt.Fprint(buf, "\n")

		if !ctx.cconfig.validator.HasWarnings() {
			break
		}

		fallthrough
	case ctx.cconfig.validator.HasWarnings():
		_, _ = fmt.Fprintf(buf, "Configuration parsed and loaded with warnings:\n\n")

		for _, err = range ctx.cconfig.validator.Warnings() {
			_, _ = fmt.Fprintf(buf, "\t - %v\n", err)
		}

		_, _ = fmt.Fprint(buf, "\n")
	default:
		_, _ = fmt.Fprintf(buf, "Configuration parsed and loaded successfully without errors.\n\n")
	}

	fmt.Print(buf.String())

	if isError {
		os.Exit(1)
	}

	return nil
}

// ConfigTemplateRunE is the RunE for the authelia validate-config command.
func (ctx *CmdCtx) ConfigTemplateRunE(_ *cobra.Command, _ []string) (err error) {
	var (
		source *configuration.FileSource
		ok     bool
	)

	buf := &bytes.Buffer{}

	var files []*configuration.File

	var n int

	for _, s := range ctx.cconfig.sources {
		if source, ok = s.(*configuration.FileSource); !ok {
			continue
		}

		n++

		if files, err = source.ReadFiles(); err != nil {
			return err
		}

		buf.WriteString(fmt.Sprintf(fmtYAMLConfigTemplateHeader, strings.Join(source.GetBytesFilterNames(), ", ")))

		for _, file := range files {
			if reYAMLComment.Match(file.Data) {
				buf.Write(reYAMLComment.ReplaceAll(file.Data, []byte(fmt.Sprintf(fmtYAMLConfigTemplateFileHeader+"$1", file.Path))))
			} else {
				buf.WriteString(fmt.Sprintf(fmtYAMLConfigTemplateFileHeader, file.Path))
				buf.Write(file.Data)
			}
		}
	}

	if n == 0 {
		return fmt.Errorf("templating requires configuration files however no configuration file sources were specified")
	}

	fmt.Println(buf.String())

	return nil
}

func newConfigValidateLegacyCmd(ctx *CmdCtx) (cmd *cobra.Command) {
	cmd = newConfigValidateCmd(ctx)

	cmd.Use = "validate-config"
	cmd.Example = cmdAutheliaConfigValidateLegacyExample

	return cmd
}
