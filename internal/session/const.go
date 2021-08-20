package session

const userSessionStorerKey = "UserSession"

const testDomain = "example.com"
const testExpiration = "40"
const testName = "my_session"
const testUsername = "john"

const (
	redisKeySeparator     = ":"
	redisKeyWildcard      = "*"
	redisKeyPrefix        = "authelia"
	redisKeyPrefixSession = "session"
	redisKeyPrefixProfile = "profile"
)
