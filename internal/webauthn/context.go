// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"context"
	"net/url"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// Context is the set of request scoped values the WebAuthn provider is built from.
type Context interface {
	context.Context

	GetOrigin() (origin *url.URL, err error)
	GetConfiguration() (config *schema.Configuration)
	GetWebAuthnMetaDataProvider() MetaDataProvider
}
