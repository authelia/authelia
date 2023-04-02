// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

// DuoAPIConfiguration represents the configuration related to Duo API.
type DuoAPIConfiguration struct {
	Disable              bool   `koanf:"disable"`
	Hostname             string `koanf:"hostname"`
	IntegrationKey       string `koanf:"integration_key"`
	SecretKey            string `koanf:"secret_key"`
	EnableSelfEnrollment bool   `koanf:"enable_self_enrollment"`
}
