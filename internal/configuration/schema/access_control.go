package schema

// AccessControlConfiguration represents the configuration related to ACLs.
type AccessControlConfiguration struct {
	DefaultPolicy string       `mapstructure:"default_policy"`
	Networks      []ACLNetwork `mapstructure:"networks"`
	Rules         []ACLRule    `mapstructure:"rules"`
}

// ACLNetwork represents one ACL network group entry; "weak" coerces a single value into slice.
type ACLNetwork struct {
	Name     []string `mapstructure:"name,weak"`
	Networks []string `mapstructure:"networks"`
}

// ACLRule represents one ACL rule entry; "weak" coerces a single value into slice.
type ACLRule struct {
	Domains   []string   `mapstructure:"domain,weak"`
	Policy    string     `mapstructure:"policy"`
	Subjects  [][]string `mapstructure:"subject,weak"`
	Networks  []string   `mapstructure:"networks"`
	Resources []string   `mapstructure:"resources"`
}
