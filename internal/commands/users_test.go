package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUsersCmdShouldRegisterSubcommands(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersCmd(ctx)

	assert.Equal(t, "users", cmd.Use)
	require.NoError(t, cmd.Args(cmd, []string{}))
	require.Error(t, cmd.Args(cmd, []string{"unexpected"}))
	assert.NotNil(t, cmd.PersistentPreRunE)

	expected := []string{"get", "list", "add", "update", "delete", "groups", "schema"}

	names := make([]string, 0, len(cmd.Commands()))

	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}

	for _, name := range expected {
		assert.Contains(t, names, name)
	}
}

func TestNewUsersGetCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersGetCmd(ctx)

	assert.Equal(t, "get <username>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	flag := cmd.Flags().Lookup(cmdFlagNameFormat)
	require.NotNil(t, flag)
	assert.Equal(t, cmdFlagValueFormatTable, flag.DefValue)

	assert.NotNil(t, cmd.Flags().Lookup(cmdFlagNameFields))

	require.NoError(t, cmd.Args(cmd, []string{"john"}))
	require.Error(t, cmd.Args(cmd, []string{}))
	require.Error(t, cmd.Args(cmd, []string{"john", "extra"}))
}

func TestNewUsersListCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersListCmd(ctx)

	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	flag := cmd.Flags().Lookup(cmdFlagNameFormat)
	require.NotNil(t, flag)
	assert.Equal(t, cmdFlagValueFormatTable, flag.DefValue)

	require.NoError(t, cmd.Args(cmd, []string{}))
	require.Error(t, cmd.Args(cmd, []string{"unexpected"}))
}

func TestNewUsersAddCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersAddCmd(ctx)

	assert.Equal(t, "add", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	flag := cmd.Flags().Lookup(cmdFlagNameFile)
	require.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)

	require.NoError(t, cmd.Args(cmd, []string{}))
	require.Error(t, cmd.Args(cmd, []string{"unexpected"}))
}

func TestNewUsersUpdateCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersUpdateCmd(ctx)

	assert.Equal(t, "update", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	require.NoError(t, cmd.Args(cmd, []string{}))
}

func TestNewUsersDeleteCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersDeleteCmd(ctx)

	assert.Equal(t, "delete <username>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	require.NoError(t, cmd.Args(cmd, []string{"john"}))
	require.Error(t, cmd.Args(cmd, []string{}))
}

func TestNewUsersGroupsCmdShouldRegisterSubcommands(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersGroupsCmd(ctx)

	assert.Equal(t, "groups", cmd.Use)

	expected := []string{"list", "add", "delete"}
	names := make([]string, 0, len(cmd.Commands()))

	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}

	for _, name := range expected {
		assert.Contains(t, names, name)
	}

	require.NoError(t, cmd.Args(cmd, []string{}))
}

func TestNewUsersGroupsListCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersGroupsListCmd(ctx)

	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	flag := cmd.Flags().Lookup(cmdFlagNameFormat)
	require.NotNil(t, flag)
	assert.Equal(t, cmdFlagValueFormatTable, flag.DefValue)
}

func TestNewUsersGroupsAddCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersGroupsAddCmd(ctx)

	assert.Equal(t, "add <group>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	require.NoError(t, cmd.Args(cmd, []string{"admins"}))
	require.Error(t, cmd.Args(cmd, []string{}))
}

func TestNewUsersGroupsDeleteCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersGroupsDeleteCmd(ctx)

	assert.Equal(t, "delete <group>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	require.NoError(t, cmd.Args(cmd, []string{"admins"}))
	require.Error(t, cmd.Args(cmd, []string{}))
}

func TestNewUsersSchemaCmd(t *testing.T) {
	ctx := NewCmdCtx()
	cmd := newUsersSchemaCmd(ctx)

	assert.Equal(t, "schema", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	require.NoError(t, cmd.Args(cmd, []string{}))
}
