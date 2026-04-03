package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewFuncs(t *testing.T) {
	var x *cobra.Command

	x = newConfigCmd(&CmdCtx{})
	assert.NotNil(t, x)

	x = newConfigValidateCmd(&CmdCtx{})
	assert.NotNil(t, x)

	x = newConfigTemplateCmd(&CmdCtx{})
	assert.NotNil(t, x)

	x = newConfigValidateLegacyCmd(&CmdCtx{})
	assert.NotNil(t, x)
}

func TestRunConfigValidate(t *testing.T) {
	errors := &schema.StructValidator{}

	errors.Push(fmt.Errorf("error one"))
	errors.Push(fmt.Errorf("error two"))

	errorsWarns := &schema.StructValidator{}

	errorsWarns.Push(fmt.Errorf("error three"))
	errorsWarns.Push(fmt.Errorf("error four"))
	errorsWarns.PushWarning(fmt.Errorf("error five"))

	warns := &schema.StructValidator{}

	warns.PushWarning(fmt.Errorf("error six"))
	warns.PushWarning(fmt.Errorf("error seven"))

	testCases := []struct {
		name      string
		validator *schema.StructValidator
		expected  string
		err       string
	}{
		{
			"ShouldHandleEmpty",
			&schema.StructValidator{},
			"Configuration parsed and loaded successfully without errors.\n\n",
			"",
		},
		{
			"ShouldHandleErrors",
			errors,
			"Configuration parsed and loaded with errors:\n\n\t - error one\n\t - error two\n\n",
			"configuration validation failed",
		},
		{
			"ShouldHandleErrorsAndWarnings",
			errorsWarns,
			"Configuration parsed and loaded with errors:\n\n\t - error three\n\t - error four\n\nConfiguration parsed and loaded with warnings:\n\n\t - error five\n\n",
			"configuration validation failed",
		},
		{
			"ShouldHandleWarnings",
			warns,
			"Configuration parsed and loaded with warnings:\n\n\t - error six\n\t - error seven\n\n",
			"",
		},
	}

	buf := new(bytes.Buffer)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := runConfigValidate(buf, tc.validator)

			assert.Equal(t, tc.expected, buf.String())

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			buf.Reset()
		})
	}
}

func TestRunConfigTemplate(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "config1.yaml")
	path2 := filepath.Join(dir, "config2.yaml")

	require.NoError(t, os.WriteFile(path1, []byte(`example:123`), 0600))
	require.NoError(t, os.WriteFile(path2, []byte("---\nexample:123"), 0600))

	testCases := []struct {
		name     string
		sources  []configuration.Source
		expected string
		err      string
	}{
		{
			"ShouldHandleNil",
			nil,
			"",
			"templating requires configuration files however no configuration file sources were specified",
		},
		{
			"ShouldHandleExample1",
			[]configuration.Source{configuration.NewFileSource(path1)},
			fmt.Sprintf("\n---\n##\n## Authelia rendered configuration file (file filters).\n##\n## Filters: \n##\n\n---\n##\n## File Source Path: %s\n##\n\nexample:123", path1),
			"",
		},
		{
			"ShouldHandleExample2",
			[]configuration.Source{configuration.NewFileSource(path2)},
			fmt.Sprintf("\n---\n##\n## Authelia rendered configuration file (file filters).\n##\n## Filters: \n##\n\n---\n##\n## File Source Path: %s\n##\n\nexample:123", path2),
			"",
		},
	}

	buf := new(bytes.Buffer)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := runConfigTemplate(buf, tc.sources)

			assert.Equal(t, tc.expected, buf.String())

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}

			buf.Reset()
		})
	}
}
