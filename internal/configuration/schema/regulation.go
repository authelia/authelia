package schema

// RegulationConfiguration represents the configuration related to regulation.
type RegulationConfiguration struct {
	MaxRetries int    `mapstructure:"max_retries"`
	FindTime   string `mapstructure:"find_time"`
	BanTime    string `mapstructure:"ban_time"`
}

var DefaultRegulationConfiguration = RegulationConfiguration{
	MaxRetries: 3,
	FindTime:   "2m",
	BanTime:    "5m",
}
