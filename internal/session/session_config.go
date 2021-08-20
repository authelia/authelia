package session

import (
	"github.com/fasthttp/session/v2"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewSessionConfig converts the schema.SessionConfiguration to a session.Config.
func NewSessionConfig(config schema.SessionConfiguration) (configuration *session.Config) {
	conf := session.NewDefaultConfig()

	// Override the cookie name.
	conf.CookieName = config.Name

	// Set the cookie to the given domain.
	conf.Domain = config.Domain

	// Set the cookie SameSite option.
	switch config.SameSite {
	case "strict":
		conf.CookieSameSite = fasthttp.CookieSameSiteStrictMode
	case "none":
		conf.CookieSameSite = fasthttp.CookieSameSiteNoneMode
	case "lax":
		conf.CookieSameSite = fasthttp.CookieSameSiteLaxMode
	default:
		conf.CookieSameSite = fasthttp.CookieSameSiteLaxMode
	}

	// Only serve the header over HTTPS.
	conf.Secure = true

	// Ignore the error as it will be handled by validator.
	conf.Expiration, _ = utils.ParseDurationString(config.Expiration)

	// TODO(conf.michaud): Make this configurable by giving the list of IPs that are trustable.
	conf.IsSecureFunc = func(*fasthttp.RequestCtx) bool {
		return true
	}

	if config.Redis != nil {
		serializer := NewEncryptingSerializer(config.Secret)

		conf.EncodeFunc = serializer.Encode
		conf.DecodeFunc = serializer.Decode
	}

	return &conf
}
