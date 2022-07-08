package authentication

// LevelToString returns a string representation of an authentication.Level.
func LevelToString(level Level) string {
	switch level {
	case NotAuthenticated:
		return "not_authenticated"
	case OneFactor:
		return "one_factor"
	case TwoFactor:
		return "two_factor"
	}

	return "invalid"
}
