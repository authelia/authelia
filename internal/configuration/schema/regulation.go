package schema

import (
	"time"
)

// Regulation represents the configuration related to regulation.
type Regulation struct {
	MaxRetries int           `koanf:"max_retries" json:"max_retries" jsonschema:"default=3,title=Maximum Retries" jsonschema_description:"The maximum number of failed attempts permitted before banning a user."`
	FindTime   time.Duration `koanf:"find_time" json:"find_time" jsonschema:"default=2 minutes,title=Find Time" jsonschema_description:"The amount of time to consider when determining the number of failed attempts."`
	BanTime    time.Duration `koanf:"ban_time" json:"ban_time" jsonschema:"default=5 minutes,title=Ban Time" jsonschema_description:"The amount of time to ban the user for when it's determined the maximum retries has been exceeded."`
}

// DefaultRegulationConfiguration represents default configuration parameters for the regulator.
var DefaultRegulationConfiguration = Regulation{
	MaxRetries: 3,
	FindTime:   time.Minute * 2,
	BanTime:    time.Minute * 5,
}
