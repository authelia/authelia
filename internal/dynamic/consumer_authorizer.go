// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package dynamic

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewAuthorizerConsumer returns a Consumer which hot reloads the access control rules of an authorization.Authorizer.
func NewAuthorizerConsumer(authorizer *authorization.Authorizer) Consumer {
	return &authorizerConsumer{authorizer: authorizer}
}

type authorizerConsumer struct {
	authorizer *authorization.Authorizer
}

// Name implements Consumer.
func (c *authorizerConsumer) Name() string {
	return "access_control"
}

// Apply implements Consumer.
func (c *authorizerConsumer) Apply(config *schema.DynamicConfiguration) (err error) {
	if c.authorizer == nil {
		return errors.New("no authorizer is available")
	}

	c.authorizer.Update(config.AccessControl)

	return nil
}
