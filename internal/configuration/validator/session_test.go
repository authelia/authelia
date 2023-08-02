package validator

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultSessionConfig() schema.Session {
	config := schema.Session{}
	config.Secret = testJWTSecret
	config.Domain = exampleDotCom //nolint:staticcheck
	config.Cookies = []schema.SessionCookie{}

	return config
}

func TestShouldSetDefaultSessionValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, schema.DefaultSessionConfiguration.Name, config.Name)
	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
	assert.Equal(t, schema.DefaultSessionConfiguration.RememberMe, config.RememberMe)
	assert.Equal(t, schema.DefaultSessionConfiguration.SameSite, config.SameSite)
}

func TestShouldSetDefaultSessionDomainsValues(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.Session
		expected schema.Session
		errs     []string
	}{
		{
			"ShouldSetGoodDefaultValues",
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					SameSite: "lax", Expiration: time.Hour, Inactivity: time.Minute, RememberMe: time.Hour * 2,
				},
				Domain: exampleDotCom,
			},
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "authelia_session", SameSite: "lax", Expiration: time.Hour, Inactivity: time.Minute, RememberMe: time.Hour * 2,
				},
				Domain: exampleDotCom,
				Cookies: []schema.SessionCookie{
					{
						SessionCookieCommon: schema.SessionCookieCommon{
							Name: "authelia_session", SameSite: "lax", Expiration: time.Hour,
							Inactivity: time.Minute, RememberMe: time.Hour * 2,
						},
						Domain: exampleDotCom,
					},
				},
			},
			nil,
		},
		{
			"ShouldNotSetBadDefaultValues",
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					SameSite: "BAD VALUE", Expiration: time.Hour, Inactivity: time.Minute, RememberMe: time.Hour * 2,
				},
				Cookies: []schema.SessionCookie{
					{
						SessionCookieCommon: schema.SessionCookieCommon{
							Name:       "authelia_session",
							Expiration: time.Hour, Inactivity: time.Minute, RememberMe: time.Hour * 2,
						},
						Domain: exampleDotCom,
					},
				},
			},
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "authelia_session", SameSite: "BAD VALUE", Expiration: time.Hour, Inactivity: time.Minute, RememberMe: time.Hour * 2,
				},
				Cookies: []schema.SessionCookie{
					{
						SessionCookieCommon: schema.SessionCookieCommon{
							Name: "authelia_session", SameSite: schema.DefaultSessionConfiguration.SameSite,
							Expiration: time.Hour, Inactivity: time.Minute, RememberMe: time.Hour * 2,
						},
						Domain: exampleDotCom,
					},
				},
			},
			[]string{
				"session: option 'same_site' must be one of 'none', 'lax', or 'strict' but it's configured as 'BAD VALUE'",
			},
		},
		{
			"ShouldSetDefaultValuesForEachConfig",
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "default_session", SameSite: "lax", Expiration: time.Hour, Inactivity: time.Minute,
					RememberMe: schema.RememberMeDisabled,
				},
				Cookies: []schema.SessionCookie{
					{
						Domain: exampleDotCom,
					},
					{
						SessionCookieCommon: schema.SessionCookieCommon{
							Name: "authelia_session", SameSite: "strict",
						},
						Domain: "example2.com",
					},
				},
			},
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "default_session", SameSite: "lax", Expiration: time.Hour, Inactivity: time.Minute,
					RememberMe: schema.RememberMeDisabled, DisableRememberMe: true,
				},
				Cookies: []schema.SessionCookie{
					{
						SessionCookieCommon: schema.SessionCookieCommon{
							Name: "default_session", SameSite: "lax",
							Expiration: time.Hour, Inactivity: time.Minute, RememberMe: schema.RememberMeDisabled, DisableRememberMe: true,
						},
						Domain: exampleDotCom,
					},
					{
						SessionCookieCommon: schema.SessionCookieCommon{
							Name: "authelia_session", SameSite: "strict",
							Expiration: time.Hour, Inactivity: time.Minute, RememberMe: schema.RememberMeDisabled, DisableRememberMe: true,
						},
						Domain: "example2.com",
					},
				},
			},
			nil,
		},
		{
			"ShouldErrorOnEmptyConfig",
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "", SameSite: "",
				},
				Domain:  "",
				Cookies: []schema.SessionCookie{},
			},
			schema.Session{
				SessionCookieCommon: schema.SessionCookieCommon{
					Name: "authelia_session", SameSite: "lax", Expiration: time.Hour, Inactivity: time.Minute * 5, RememberMe: time.Hour * 24 * 30,
				},
				Cookies: []schema.SessionCookie{},
			},
			[]string{
				"session: option 'cookies' is required",
			},
		},
	}

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator.Clear()

			have := tc.have

			ValidateSession(&have, validator)

			assert.Len(t, validator.Warnings(), 0)

			errs := validator.Errors()
			require.Len(t, validator.Errors(), len(tc.errs))

			for i, err := range errs {
				assert.EqualError(t, err, tc.errs[i])
			}

			assert.Equal(t, tc.expected, have)
		})
	}
}

func TestShouldSetDefaultSessionValuesWhenNegative(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Expiration, config.Inactivity, config.RememberMe = -1, -1, -2

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
	assert.Equal(t, schema.DefaultSessionConfiguration.RememberMe, config.RememberMe)
}

func TestShouldWarnSessionValuesWhenPotentiallyInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Domain = ".example.com" //nolint:staticcheck

	ValidateSession(&config, validator)

	require.Len(t, validator.Warnings(), 1)
	assert.Len(t, validator.Errors(), 0)

	assert.EqualError(t, validator.Warnings()[0], "session: domain config #1 (domain '.example.com'): option 'domain' has a prefix of '.' which is not supported or intended behaviour: you can use this at your own risk but we recommend removing it")
}

func TestShouldHandleRedisConfigSuccessfully(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	config = newDefaultSessionConfig()

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.SessionRedis{
		Host:     "redis.localhost",
		Port:     6379,
		Password: "password",
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, 8, config.Redis.MaximumActiveConnections)
}

func TestShouldRaiseErrorWithInvalidRedisPortLow(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "authelia-port-1",
		Port: -1,
	}

	ValidateSession(&config, validator)

	require.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, -1))
}

func TestShouldRaiseErrorWithInvalidRedisPortHigh(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "authelia-port-1",
		Port: 65536,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, 65536))
}

func TestShouldRaiseErrorWhenRedisIsUsedAndSecretNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Secret = ""

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	config = newDefaultSessionConfig()
	config.Secret = ""

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.SessionRedis{
		Host: "redis.localhost",
		Port: 6379,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionSecretRequired, "redis"))
}

func TestShouldRaiseErrorWhenRedisHasHostnameButNoPort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	config = newDefaultSessionConfig()

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.SessionRedis{
		Host: "redis.localhost",
		Port: 0,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "session: redis: option 'port' must be between 1 and 65535 but it's configured as '0'")
}

func TestShouldRaiseOneErrorWhenRedisHighAvailabilityHasNodesWithNoHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "redis",
		Port: 6379,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelName:     "authelia-sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.SessionRedisHighAvailabilityNode{
				{
					Port: 26379,
				},
				{
					Port: 26379,
				},
			},
		},
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 1)

	assert.EqualError(t, errors[0], "session: redis: high_availability: option 'nodes': option 'host' is required for each node but one or more nodes are missing this")
}

func TestShouldRaiseOneErrorWhenRedisHighAvailabilityDoesNotHaveSentinelName(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "redis",
		Port: 6379,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelPassword: "abc123",
		},
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 1)

	assert.EqualError(t, errors[0], "session: redis: high_availability: option 'sentinel_name' is required")
}

func TestShouldUpdateDefaultPortWhenRedisSentinelHasNodes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "redis",
		Port: 6379,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelName:     "authelia-sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.SessionRedisHighAvailabilityNode{
				{
					Host: "node-1",
					Port: 333,
				},
				{
					Host: "node-2",
				},
				{
					Host: "node-3",
				},
			},
		},
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())

	assert.Equal(t, 333, config.Redis.HighAvailability.Nodes[0].Port)
	assert.Equal(t, 26379, config.Redis.HighAvailability.Nodes[1].Port)
	assert.Equal(t, 26379, config.Redis.HighAvailability.Nodes[2].Port)
}

func TestShouldRaiseErrorsWhenRedisSentinelOptionsIncorrectlyConfigured(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Secret = ""
	config.Redis = &schema.SessionRedis{
		Port: 65536,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.SessionRedisHighAvailabilityNode{
				{
					Host: "node1",
					Port: 26379,
				},
			},
			RouteByLatency: true,
			RouteRandomly:  true,
		},
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 2)

	assert.EqualError(t, errors[0], fmt.Sprintf(errFmtSessionRedisPortRange, 65536))
	assert.EqualError(t, errors[1], fmt.Sprintf(errFmtSessionSecretRequired, "redis"))

	validator.Clear()

	config = newDefaultSessionConfig()

	config.Secret = ""
	config.Redis = &schema.SessionRedis{
		Port: -1,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.SessionRedisHighAvailabilityNode{
				{
					Host: "node1",
					Port: 26379,
				},
			},
			RouteByLatency: true,
			RouteRandomly:  true,
		},
	}

	ValidateSession(&config, validator)

	errors = validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 2)

	assert.EqualError(t, errors[0], fmt.Sprintf(errFmtSessionRedisPortRange, -1))
	assert.EqualError(t, errors[1], fmt.Sprintf(errFmtSessionSecretRequired, "redis"))
}

func TestShouldNotRaiseErrorsAndSetDefaultPortWhenRedisSentinelPortBlank(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "mysentinelHost",
		Port: 0,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.SessionRedisHighAvailabilityNode{
				{
					Host: "node1",
					Port: 26379,
				},
			},
			RouteByLatency: true,
			RouteRandomly:  true,
		},
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())

	assert.Equal(t, 26379, config.Redis.Port)
}

func TestShouldRaiseErrorWhenRedisHostAndHighAvailabilityNodesEmpty(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Port: 26379,
		HighAvailability: &schema.SessionRedisHighAvailability{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			RouteByLatency:   true,
			RouteRandomly:    true,
		},
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], errFmtSessionRedisHostOrNodesRequired)
}

func TestShouldRaiseErrorsWhenRedisHostNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Port: 6379,
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 1)

	assert.EqualError(t, errors[0], errFmtSessionRedisHostRequired)
}

func TestShouldSetDefaultRedisTLSOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "redis.local",
		Port: 6379,
		TLS:  &schema.TLS{},
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, uint16(tls.VersionTLS12), config.Redis.TLS.MinimumVersion.Value)
	assert.Equal(t, uint16(0), config.Redis.TLS.MaximumVersion.Value)
	assert.Equal(t, "redis.local", config.Redis.TLS.ServerName)
}

func TestShouldRaiseErrorOnBadRedisTLSOptionsSSL30(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "redis.local",
		Port: 6379,
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
		},
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: redis: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func TestShouldRaiseErrorOnBadRedisTLSOptionsMinVerGreaterThanMax(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.SessionRedis{
		Host: "redis.local",
		Port: 6379,
		TLS: &schema.TLS{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
			MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
		},
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: redis: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.0")
}

func TestShouldRaiseErrorWhenHaveDuplicatedDomainName(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = "" //nolint:staticcheck
	config.Cookies = append(config.Cookies, schema.SessionCookie{
		Domain:      exampleDotCom,
		AutheliaURL: MustParseURL("https://login.example.com"),
	})
	config.Cookies = append(config.Cookies, schema.SessionCookie{
		Domain:      exampleDotCom,
		AutheliaURL: MustParseURL("https://login.example.com"),
	})

	ValidateSession(&config, validator)
	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionDomainDuplicate, sessionDomainDescriptor(1, schema.SessionCookie{Domain: exampleDotCom})))
}

func TestShouldRaiseErrorWhenSubdomainConflicts(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = "" //nolint:staticcheck
	config.Cookies = append(config.Cookies, schema.SessionCookie{
		Domain:      exampleDotCom,
		AutheliaURL: MustParseURL("https://login.example.com"),
	})
	config.Cookies = append(config.Cookies, schema.SessionCookie{
		Domain:      "internal.example.com",
		AutheliaURL: MustParseURL("https://login.internal.example.com"),
	})

	ValidateSession(&config, validator)
	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "session: domain config #2 (domain 'internal.example.com'): option 'domain' shares the same cookie domain scope as another configured session domain")
}

func TestShouldRaiseErrorWhenDomainIsInvalid(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		warnings []string
		expected []string
	}{
		{"ShouldNotRaiseErrorOnValidDomain", exampleDotCom, nil, nil},
		{"ShouldNotRaiseErrorOnValidIPLocalHost1", "127.0.0.1", nil, nil},
		{"ShouldNotRaiseErrorOnValidIPLocalHost30", "127.0.0.30", nil, nil},
		{"ShouldNotRaiseErrorOnValidIPClassC40", "192.168.0.40", nil, nil},
		{"ShouldRaiseErrorOnMissingDomain", "", nil, []string{"session: domain config #1 (domain ''): option 'domain' is required"}},
		{"ShouldRaiseErrorOnDomainWithInvalidChars", "example!.com", nil, []string{"session: domain config #1 (domain 'example!.com'): option 'domain' does not appear to be a valid cookie domain or an ip address"}},
		{"ShouldRaiseErrorOnDomainWithoutDots", "localhost", nil, []string{"session: domain config #1 (domain 'localhost'): option 'domain' is not a valid cookie domain: must have at least a single period or be an ip address"}},
		{"ShouldRaiseErrorOnPublicDomainDuckDNS", "duckdns.org", nil, []string{"session: domain config #1 (domain 'duckdns.org'): option 'domain' is not a valid cookie domain: the domain is part of the special public suffix list"}},
		{"ShouldNotRaiseErrorOnSuffixOfPublicDomainDuckDNS", "example.duckdns.org", nil, nil},
		{"ShouldRaiseWarningOnDomainWithLeadingDot", ".example.com", []string{"session: domain config #1 (domain '.example.com'): option 'domain' has a prefix of '.' which is not supported or intended behaviour: you can use this at your own risk but we recommend removing it"}, nil},
		{"ShouldRaiseErrorOnDomainWithLeadingStarDot", "*.example.com", nil, []string{"session: domain config #1 (domain '*.example.com'): option 'domain' must be the domain you wish to protect not a wildcard domain but it's configured as '*.example.com'"}},
		{"ShouldRaiseErrorOnDomainNotSet", "", nil, []string{"session: domain config #1 (domain ''): option 'domain' is required"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := newDefaultSessionConfig()
			config.Domain = "" //nolint:staticcheck

			config.Cookies = []schema.SessionCookie{
				{
					Domain: tc.have,
				},
			}

			ValidateSession(&config, validator)

			require.Len(t, validator.Warnings(), len(tc.warnings))
			require.Len(t, validator.Errors(), len(tc.expected))

			for i, expected := range tc.warnings {
				assert.EqualError(t, validator.Warnings()[i], expected)
			}

			for i, expected := range tc.expected {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}
}

func TestShouldRaiseErrorWhenPortalURLIsInvalid(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected []string
	}{
		{"ShouldRaiseErrorOnInvalidScope", "https://example2.com/login", []string{"session: domain config #1 (domain 'example.com'): option 'authelia_url' does not share a cookie scope with domain 'example.com' with a value of 'https://example2.com/login'"}},
		{"ShouldRaiseErrorOnInvalidScheme", "http://example.com/login", []string{"session: domain config #1 (domain 'example.com'): option 'authelia_url' does not have a secure scheme with a value of 'http://example.com/login'"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := newDefaultSessionConfig()
			config.Domain = "" //nolint:staticcheck
			config.Cookies = []schema.SessionCookie{
				{
					SessionCookieCommon: schema.SessionCookieCommon{
						Name: "authelia_session",
					},
					Domain:      exampleDotCom,
					AutheliaURL: MustParseURL(tc.have)},
			}

			ValidateSession(&config, validator)

			assert.Len(t, validator.Warnings(), 0)
			require.Len(t, validator.Errors(), len(tc.expected))

			for i, expected := range tc.expected {
				assert.EqualError(t, validator.Errors()[i], expected)
			}
		})
	}
}

func TestShouldRaiseErrorWhenSameSiteSetIncorrectly(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.SameSite = "NOne"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "session: option 'same_site' must be one of 'none', 'lax', or 'strict' but it's configured as 'NOne'")
	assert.EqualError(t, validator.Errors()[1], "session: domain config #1 (domain 'example.com'): option 'same_site' must be one of 'none', 'lax', or 'strict' but it's configured as 'NOne'")
}

func TestShouldNotRaiseErrorWhenSameSiteSetCorrectly(t *testing.T) {
	validator := schema.NewStructValidator()

	var config schema.Session

	validOptions := []string{"none", "lax", "strict"}

	for _, opt := range validOptions {
		validator.Clear()

		config = newDefaultSessionConfig()
		config.SameSite = opt

		ValidateSession(&config, validator)

		assert.Len(t, validator.Warnings(), 0)
		assert.Len(t, validator.Errors(), 0)
	}
}

func TestShouldSetDefaultWhenNegativeAndNotOverrideDisabledRememberMe(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Inactivity = -1
	config.Expiration = -1
	config.RememberMe = schema.RememberMeDisabled

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
	assert.Equal(t, schema.RememberMeDisabled, config.RememberMe)
	assert.True(t, config.DisableRememberMe)
}

func TestShouldSetDefaultRememberMeDuration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, config.RememberMe, schema.DefaultSessionConfiguration.RememberMe)
}

func TestShouldNotAllowLegacyAndModernCookiesConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Cookies = append(config.Cookies, schema.SessionCookie{
		SessionCookieCommon: schema.SessionCookieCommon{
			Name:       config.Name,
			SameSite:   config.SameSite,
			Expiration: config.Expiration,
			Inactivity: config.Inactivity,
			RememberMe: config.RememberMe,
		},
		Domain: config.Domain, //nolint:staticcheck
	})

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: option 'domain' and option 'cookies' can't be specified at the same time")
}

func MustParseURL(uri string) *url.URL {
	u, err := url.ParseRequestURI(uri)

	if err != nil {
		panic(err)
	}

	return u
}
