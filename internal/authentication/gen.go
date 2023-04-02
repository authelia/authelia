// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/authentication`.

//go:generate mockgen -package authentication -destination ldap_client_mock.go -mock_names LDAPClient=MockLDAPClient github.com/authelia/authelia/v4/internal/authentication LDAPClient
//go:generate mockgen -package authentication -destination ldap_client_factory_mock.go -mock_names LDAPClientFactory=MockLDAPClientFactory github.com/authelia/authelia/v4/internal/authentication LDAPClientFactory
