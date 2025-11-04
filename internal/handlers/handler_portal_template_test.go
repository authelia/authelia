package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/mocks"
)

type okResponse struct {
	Status string                          `json:"status"`
	Data   portalTemplateConfigurationBody `json:"data"`
}

func TestPortalTemplateDefault(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	PortalTemplateGET(mock.Ctx)

	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	var body okResponse
	assert.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &body))
	assert.Equal(t, "OK", body.Status)
	assert.Equal(t, "none", body.Data.Template)
	assert.False(t, body.Data.EnableTemplateSwitcher)
}

func TestPortalTemplateConfigured(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Configuration.PortalTemplate = "Gateway"
	mock.Ctx.Configuration.PortalTemplateSwitcher = true

	PortalTemplateGET(mock.Ctx)

	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	var body okResponse
	assert.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &body))
	assert.Equal(t, "OK", body.Status)
	assert.Equal(t, "gateway", body.Data.Template)
	assert.True(t, body.Data.EnableTemplateSwitcher)
}
