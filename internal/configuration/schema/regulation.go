// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"time"
)

// RegulationConfiguration represents the configuration related to regulation.
type RegulationConfiguration struct {
	MaxRetries int           `koanf:"max_retries"`
	FindTime   time.Duration `koanf:"find_time,weak"`
	BanTime    time.Duration `koanf:"ban_time,weak"`
}

// DefaultRegulationConfiguration represents default configuration parameters for the regulator.
var DefaultRegulationConfiguration = RegulationConfiguration{
	MaxRetries: 3,
	FindTime:   time.Minute * 2,
	BanTime:    time.Minute * 5,
}
