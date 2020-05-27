package suites

// GetLoginBaseURL returns the URL of the login portal and the path prefix if specified.
func GetLoginBaseURL() string {
	if PathPrefix != "" {
		return LoginBaseURL + PathPrefix
	}

	return LoginBaseURL
}
