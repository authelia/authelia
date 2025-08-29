package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewDebugCmds(t *testing.T) {
	var cmd *cobra.Command

	cmd = newDebugCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugExpressionCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugOIDCCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugOIDCClaimsCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newDebugTLSCmd(&CmdCtx{})
	assert.NotNil(t, cmd)
}
