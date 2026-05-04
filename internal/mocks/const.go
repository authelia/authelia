package mocks

const (
	cookieNameAutheliaSession = "authelia_session"

	schemeHTTPS = "https"

	policyBypass    = "bypass"
	policyOneFactor = "one_factor"
	policyTwoFactor = "two_factor"
	policyDeny      = "deny"

	subjectGroupAdmin   = "group:admin"
	subjectGroupGrafana = "group:grafana"
)

const (
	domainBypassGet     = "bypass-get.example.com"
	domainBypassHead    = "bypass-head.example.com"
	domainBypassOptions = "bypass-options.example.com"
	domainBypassTrace   = "bypass-trace.example.com" //nolint:gosec
	domainBypassPut     = "bypass-put.example.com"   //nolint:gosec
	domainBypassPatch   = "bypass-patch.example.com" //nolint:gosec
	domainBypassPost    = "bypass-post.example.com"
	domainBypassDelete  = "bypass-delete.example.com"
	domainBypassConnect = "bypass-connect.example.com"
)
