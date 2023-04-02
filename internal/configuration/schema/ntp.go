// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"time"
)

// NTPConfiguration represents the configuration related to ntp server.
type NTPConfiguration struct {
	Address             string        `koanf:"address"`
	Version             int           `koanf:"version"`
	MaximumDesync       time.Duration `koanf:"max_desync"`
	DisableStartupCheck bool          `koanf:"disable_startup_check"`
	DisableFailure      bool          `koanf:"disable_failure"`
}

// DefaultNTPConfiguration represents default configuration parameters for the NTP server.
var DefaultNTPConfiguration = NTPConfiguration{
	Address:       "time.cloudflare.com:123",
	Version:       4,
	MaximumDesync: time.Second * 3,
}
