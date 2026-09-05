package webauthn

import (
	"context"
	"net/url"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

type Context interface {
	context.Context

	GetOrigin() (origin *url.URL, err error)
	GetConfiguration() (config *schema.Configuration)
	GetWebAuthnMetaDataProvider() MetaDataProvider
}
