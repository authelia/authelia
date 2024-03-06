package schema

// AccessControl represents the configuration related to ACLs.
type AccessControl struct {
	// The default policy if no other policy matches the request.
	DefaultPolicy string `koanf:"default_policy" json:"default_policy" jsonschema:"default=deny,enum=deny,enum=one_factor,enum=two_factor,title=Default Authorization Policy" jsonschema_description:"The default policy applied to all authorization requests unrelated to OpenID Connect 1.0."`

	// Represents a list of named network groups.
	Networks []AccessControlNetwork `koanf:"networks" json:"networks" jsonschema:"title=Named Networks" jsonschema_description:"The list of named networks which can be reused in any ACL rule."`

	// The ACL rules list.
	Rules []AccessControlRule `koanf:"rules" json:"rules" jsonschema:"title=Rules List" jsonschema_description:"The list of ACL rules to enumerate for requests."`
}

// AccessControlNetwork represents one ACL network group entry.
type AccessControlNetwork struct {
	Name     string                       `koanf:"name" json:"name" jsonschema:"required,title=Network Name" jsonschema_description:"The name of this network to be used in the networks section of the rules section."`
	Networks AccessControlNetworkNetworks `koanf:"networks" json:"networks" jsonschema:"required,title=Networks" jsonschema_description:"The remote IP's or network ranges in CIDR notation that this rule applies to."`
}

// AccessControlRule represents one ACL rule entry.
type AccessControlRule struct {
	Domains      AccessControlRuleDomains   `koanf:"domain" json:"domain" jsonschema:"oneof_required=Domain,uniqueItems,title=Domain Literals" jsonschema_description:"The literal domains to match the domain against that this rule applies to."`
	DomainsRegex AccessControlRuleRegex     `koanf:"domain_regex" json:"domain_regex" jsonschema:"oneof_required=Domain Regex,title=Domain Regex Patterns" jsonschema_description:"The regex patterns to match the domain against that this rule applies to."`
	Policy       string                     `koanf:"policy" json:"policy" jsonschema:"required,enum=bypass,enum=deny,enum=one_factor,enum=two_factor,title=Rule Policy" jsonschema_description:"The policy this rule applies when all criteria match."`
	Subjects     AccessControlRuleSubjects  `koanf:"subject" json:"subject" jsonschema:"title=AccessControlRuleSubjects" jsonschema_description:"The users or groups that this rule applies to."`
	Networks     AccessControlRuleNetworks  `koanf:"networks" json:"networks" jsonschema:"title=Networks" jsonschema_description:"The remote IP's, network ranges in CIDR notation, or network names that this rule applies to."`
	Resources    AccessControlRuleRegex     `koanf:"resources" json:"resources" jsonschema:"title=Resources or Paths" jsonschema_description:"The regex patterns to match the resource paths that this rule applies to."`
	Methods      AccessControlRuleMethods   `koanf:"methods" json:"methods" jsonschema:"enum=GET,enum=HEAD,enum=POST,enum=PUT,enum=DELETE,enum=CONNECT,enum=OPTIONS,enum=TRACE,enum=PATCH,enum=PROPFIND,enum=PROPPATCH,enum=MKCOL,enum=COPY,enum=MOVE,enum=LOCK,enum=UNLOCK" jsonschema_description:"The list of request methods this rule applies to."`
	Query        [][]AccessControlRuleQuery `koanf:"query" json:"query" jsonschema:"title=Query Rules" jsonschema_description:"The list of query parameter rules this rule applies to."`
}

// AccessControlRuleQuery represents the ACL query criteria.
type AccessControlRuleQuery struct {
	Operator string `koanf:"operator" json:"operator" jsonschema:"enum=equal,enum=not equal,enum=present,enum=absent,enum=pattern,enum=not pattern,title=Operator" jsonschema_description:"The list of query parameter rules this rule applies to."`
	Key      string `koanf:"key" json:"key" jsonschema:"required,title=Key" jsonschema_description:"The Query Parameter key this rule applies to."`
	Value    any    `koanf:"value" json:"value" jsonschema:"title=Value" jsonschema_description:"The Query Parameter value for this rule."`
}

// DefaultACLNetwork represents the default configuration related to access control network group configuration.
var DefaultACLNetwork = []AccessControlNetwork{
	{
		Name:     "localhost",
		Networks: []string{"127.0.0.1"},
	},
	{
		Name:     "internal",
		Networks: []string{"10.0.0.0/8"},
	},
}

// DefaultACLRule represents the default configuration related to access control rule configuration.
var DefaultACLRule = []AccessControlRule{
	{
		Domains: []string{"public.example.com"},
		Policy:  "bypass",
	},
	{
		Domains: []string{"singlefactor.example.com"},
		Policy:  "one_factor",
	},
	{
		Domains: []string{"secure.example.com"},
		Policy:  policyTwoFactor,
	},
}
