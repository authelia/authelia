package schema

// RegulationConfiguration represents the configuration related to regulation.
type RegulationConfiguration struct {
	MaxRetries int   `yaml:"max_retries"`
	FindTime   int64 `yaml:"find_time"`
	BanTime    int64 `yaml:"ban_time"`
}
