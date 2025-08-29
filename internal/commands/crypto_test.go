package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewCrypto(t *testing.T) {
	var cmd *cobra.Command

	cmd = newCryptoCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newCryptoCertificateCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newCryptoPairCmd(&CmdCtx{})
	assert.NotNil(t, cmd)

	cmd = newCryptoPairSubCmd(&CmdCtx{}, "generate")
	assert.NotNil(t, cmd)

	cmd = newCryptoPairSubCmd(&CmdCtx{}, "verify")
	assert.NotNil(t, cmd)
}
