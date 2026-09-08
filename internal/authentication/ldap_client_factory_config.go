package authentication

// LDAPClientFactoryOptions represents the options used when obtaining an LDAP client.
type LDAPClientFactoryOptions struct {
	Address  string
	Username string
	Password string

	PermitUnauthenticatedBind bool
}

// LDAPClientFactoryOption is a function which configures the LDAPClientFactoryOptions.
type LDAPClientFactoryOption func(*LDAPClientFactoryOptions)

// WithAddress returns an LDAPClientFactoryOption which sets the address.
func WithAddress(address string) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.Address = address
	}
}

// WithUsername returns an LDAPClientFactoryOption which sets the username.
func WithUsername(username string) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.Username = username
	}
}

// WithPassword returns an LDAPClientFactoryOption which sets the password.
func WithPassword(password string) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.Password = password
	}
}

// WithPermitUnauthenticatedBind returns an LDAPClientFactoryOption which sets the permit unauthenticated bind option.
func WithPermitUnauthenticatedBind(permit bool) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.PermitUnauthenticatedBind = permit
	}
}
