// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package mocks

// This file is used to generate mocks. You can generate all mocks using the
// command `go generate github.com/authelia/authelia/v4/internal/mocks`.

//go:generate mockgen -package mocks -destination user_provider.go -mock_names UserProvider=MockUserProvider github.com/authelia/authelia/v4/internal/authentication UserProvider
//go:generate mockgen -package mocks -destination notifier.go -mock_names Notifier=MockNotifier github.com/authelia/authelia/v4/internal/notification Notifier
//go:generate mockgen -package mocks -destination totp.go -mock_names Provider=MockTOTP github.com/authelia/authelia/v4/internal/totp Provider
//go:generate mockgen -package mocks -destination storage.go -mock_names Provider=MockStorage github.com/authelia/authelia/v4/internal/storage Provider
//go:generate mockgen -package mocks -destination duo_api.go -mock_names API=MockAPI github.com/authelia/authelia/v4/internal/duo API
//go:generate mockgen -package mocks -destination random.go -mock_names Provider=MockRandom github.com/authelia/authelia/v4/internal/random Provider
