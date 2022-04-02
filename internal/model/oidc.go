package model

// OIDCWorkflowSession represent an OIDC workflow session.
type OIDCWorkflowSession struct {
	ClientID          string
	RequestedScopes   []string
	GrantedScopes     []string
	RequestedAudience []string
	GrantedAudience   []string
	TargetURI         string
	AuthURI           string
	Require2FA        bool
	CreatedTimestamp  int64
}
