package schema

import (
	"time"
)

// Regulation represents the configuration related to regulation.
type Regulation struct {
	Modes      []string      `koanf:"modes" yaml:"modes,omitempty" toml:"modes,omitempty" json:"modes,omitempty" jsonschema:"default=user,enum=user,enum=ip,title=Regulation Modes" jsonschema_description:"The modes to use for regulation."`
	MaxRetries int           `koanf:"max_retries" yaml:"max_retries" toml:"max_retries" json:"max_retries" jsonschema:"default=3,title=Maximum Retries" jsonschema_description:"The maximum number of failed attempts permitted before banning a user."`
	FindTime   time.Duration `koanf:"find_time" yaml:"find_time,omitempty" toml:"find_time,omitempty" json:"find_time,omitempty" jsonschema:"default=2 minutes,title=Find Time" jsonschema_description:"The amount of time to consider when determining the number of failed attempts."`
	BanTime    time.Duration `koanf:"ban_time" yaml:"ban_time,omitempty" toml:"ban_time,omitempty" json:"ban_time,omitempty" jsonschema:"default=5 minutes,title=Ban Time" jsonschema_description:"The amount of time to ban the user for when it's determined the maximum retries has been exceeded."`
}

// DefaultRegulationConfiguration represents default configuration parameters for the regulator.
var DefaultRegulationConfiguration = Regulation{
	Modes:      []string{"user"},
	MaxRetries: 3,
	FindTime:   time.Minute * 2,
	BanTime:    time.Minute * 5,
}
