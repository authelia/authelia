package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestGetStorageProvider(t *testing.T) {
	assert.Nil(t, getStorageProvider(NewCmdCtx()))
}

func TestGetAuthenticationProvider(t *testing.T) {
	var (
		ctxEmpty *CmdCtx = NewCmdCtx()
		ctxFile  *CmdCtx = NewCmdCtx()
		ctxLDAP  *CmdCtx = NewCmdCtx()
		ctxDB    *CmdCtx = NewCmdCtx()
	)

	ctxFile.config.AuthenticationBackend.File = &schema.AuthenticationBackendFile{}
	ctxLDAP.config.AuthenticationBackend.LDAP = &schema.AuthenticationBackendLDAP{}
	ctxDB.config.AuthenticationBackend.DB = &schema.AuthenticationBackendDB{}

	assert.Nil(t, getAuthenticationProvider(ctxEmpty))
	assert.IsType(t, &authentication.FileUserProvider{}, getAuthenticationProvider(ctxFile))
	assert.IsType(t, &authentication.LDAPUserProvider{}, getAuthenticationProvider(ctxLDAP))
	assert.IsType(t, &authentication.DBUserProvider{}, getAuthenticationProvider(ctxDB))
}
