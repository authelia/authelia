// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"time"
)

const (
	testDomain                   = "example.com"
	testExpiration               = time.Second * 40
	testRememberMe               = time.Hour * 24
	testAnonymousExpiration      = time.Second * 20
	testName                     = "my_session"
	testNameAuthz                = "__Secure-my_session_authz"
	testNameSession              = "__Host-Http-my_session"
	testNameCSRF                 = "__Host-my_session_csrf"
	testNameAuthorizationBinding = "__Host-Http-my_session_oauth2"
	testUsername                 = "john"
	testSecret                   = "a-secret-value"
	testHMACKey                  = "an-hmac-key"
	testHMACKeyAuthz             = "an-authz-hmac-key"
	testHMACKeyCSRF              = "a-csrf-hmac-key"
)

var errTestFailure = errors.New("bad stuff")
