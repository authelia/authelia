// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"time"
)

const (
	testDomain     = "example.com"
	testExpiration = time.Second * 40
	testName       = "my_session"
	testUsername   = "john"
)

const (
	userSessionStorerKey = "UserSession"
	randomSessionChars   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!#$%^*"
)
