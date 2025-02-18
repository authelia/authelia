package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckSchemaVersionShouldFailIfThereNotStoragePovider(t *testing.T) {
	var ctx *CmdCtx = NewCmdCtx()

	assert.ErrorContains(t, ctx.CheckSchemaVersion(), "storage not loaded")
}
