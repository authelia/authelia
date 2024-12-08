package authentication

type LDAPClientFactoryOptions struct {
	Address  string
	Username string
	Password string
}

type LDAPClientFactoryOption func(*LDAPClientFactoryOptions)

func WithAddress(address string) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.Address = address
	}
}

func WithUsername(username string) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.Username = username
	}
}

func WithPassword(password string) func(*LDAPClientFactoryOptions) {
	return func(settings *LDAPClientFactoryOptions) {
		settings.Password = password
	}
}
