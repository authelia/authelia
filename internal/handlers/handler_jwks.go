// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// JSONWebKeySetGET returns the JSON Web Key Set. Used in OAuth 2.0 and OpenID Connect 1.0.
func JSONWebKeySetGET(ctx *middlewares.AutheliaCtx) {
	ctx.SetContentType("application/json")

	if err := json.NewEncoder(ctx).Encode(ctx.Providers.OpenIDConnect.KeyManager.GetKeySet()); err != nil {
		ctx.Error(err, "failed to serve json web key set")
	}
}
