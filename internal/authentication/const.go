package authentication

// Level is the type representing a level of authentication
type Level int

const (
	// NotAuthenticated if the user is not authenticated yet.
	NotAuthenticated Level = iota
	// OneFactor if the user has passed first factor only.
	OneFactor Level = iota
	// TwoFactor if the user has passed two factors.
	TwoFactor Level = iota
)

const (
	// TOTP Method using Time-Based One-Time Password applications like Google Authenticator
	TOTP = "totp"
	// U2F Method using U2F devices like Yubikeys
	U2F = "u2f"
	// DuoPush Method using Duo application to receive push notifications.
	DuoPush = "duo_push"
)

// PossibleMethods is the set of all possible 2FA methods.
var PossibleMethods = []string{TOTP, U2F, DuoPush}
