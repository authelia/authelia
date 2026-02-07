package server

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

const (
	assetsOpenAPIPath = "public_html/api/openapi.yml"
	localOpenAPIPath  = "../../api/openapi.yml"
)

type ReadFileOpenAPI struct{}

func (lfs *ReadFileOpenAPI) Open(name string) (fs.File, error) {
	switch name {
	case assetsOpenAPIPath:
		return os.Open(localOpenAPIPath)
	default:
		return assets.Open(name)
	}
}

func (lfs *ReadFileOpenAPI) ReadFile(name string) ([]byte, error) {
	switch name {
	case assetsOpenAPIPath:
		return os.ReadFile(localOpenAPIPath)
	default:
		return assets.ReadFile(name)
	}
}

func TestShouldTemplateOpenAPI(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	fs := &ReadFileOpenAPI{}

	require.NoError(t, provider.LoadTemplatedAssets(fs))

	mock := mocks.NewMockAutheliaCtx(t)

	mock.Ctx.Configuration.Server = schema.DefaultServerConfiguration
	mock.Ctx.Configuration.Session = schema.Session{
		Cookies: []schema.SessionCookie{
			{
				Domain: "example.com",
			},
		},
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil, nil)

	opts := NewTemplatedFileOptions(&mock.Ctx.Configuration)

	handler := ServeTemplatedOpenAPI(provider.GetAssetOpenAPISpecTemplate(), opts)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "auth.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/api/openapi.yml")

	handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	body := string(mock.Ctx.Response.Body())

	assert.NotEqual(t, "", body)
	assert.Contains(t, body, "example: 'https://auth.example.com/?rd=https%3A%2F%2Fexample.com%2F&rm=GET'")
}
