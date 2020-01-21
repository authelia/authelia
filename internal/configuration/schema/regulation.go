package schema

// RegulationConfiguration represents the configuration related to regulation.
type RegulationConfiguration struct {
	MaxRetries int   `mapstructure:"max_retries"`
	FindTime   int64 `mapstructure:"find_time"`
	BanTime    int64 `mapstructure:"ban_time"`
}
