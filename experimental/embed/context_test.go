package embed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := &ctxEmbed{}

	assert.Nil(t, ctx.GetConfiguration())
	assert.Nil(t, ctx.GetLogger())

	providers := ctx.GetProviders()

	assert.Nil(t, providers.StorageProvider)
	assert.Nil(t, providers.Notifier)
	assert.Nil(t, providers.UserProvider)
	assert.Nil(t, providers.SessionProvider)
	assert.Nil(t, providers.MetaDataService)
	assert.Nil(t, providers.Metrics)
	assert.Nil(t, providers.Templates)
	assert.Nil(t, providers.Random)
	assert.Nil(t, providers.OpenIDConnect)
	assert.Nil(t, providers.UserAttributeResolver)
	assert.Nil(t, providers.Authorizer)
	assert.Nil(t, providers.NTP)
	assert.Nil(t, providers.TOTP)
}
