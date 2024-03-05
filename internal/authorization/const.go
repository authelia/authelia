package authorization

// Level is the type representing an authorization level.
type Level int

const (
	// Bypass bypass level.
	Bypass Level = iota

	// OneFactor one factor level.
	OneFactor

	// TwoFactor two factor level.
	TwoFactor

	// Denied denied level.
	Denied
)

const (
	prefixUser         = "user:"
	prefixGroup        = "group:"
	prefixOAuth2Client = "oauth2:client:"
)

const (
	lenPrefixUser         = len(prefixUser)
	lenPrefixGroup        = len(prefixGroup)
	lenPrefixOAuth2Client = len(prefixOAuth2Client)
)

const (
	bypass    = "bypass"
	oneFactor = "one_factor"
	twoFactor = "two_factor"
	deny      = "deny"
)

const (
	operatorPresent    = "present"
	operatorAbsent     = "absent"
	operatorEqual      = "equal"
	operatorNotEqual   = "not equal"
	operatorPattern    = "pattern"
	operatorNotPattern = "not pattern"
)

const (
	subexpNameUser  = "User"
	subexpNameGroup = "Group"
)

var (
	// IdentitySubexpNames is a list of valid regex subexp names.
	IdentitySubexpNames = []string{subexpNameUser, subexpNameGroup}
)

const traceFmtACLHitMiss = "ACL %s Position %d for subject %s and object %s (method %s, policy %s)"
