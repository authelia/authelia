package validator

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/knadh/koanf"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func SecretEnvNameReplacer(name string) (result string) {
	for _, secretKey := range SecretNames {
		badKey := strings.ReplaceAll(secretKey, "_", ".")

		if name == badKey {
			return secretKey
		}
	}

	return name
}

// SecretNameToEnvName converts a secret name into the env name.
func SecretNameToEnvName(secretName string) (envName string) {
	return "authelia." + secretName + ".file"
}

// IsSecretKey returns true if the key is the name of a secret.
func IsSecretKey(value string) (isSecretKey bool) {
	for _, secretKey := range SecretNames {
		if value == secretKey || value == SecretNameToEnvName(secretKey) {
			return true
		}
	}

	return false
}

// ValidateSecrets checks that secrets are either specified by config file/env or by file references.
func ValidateSecrets(configuration *schema.Configuration, validator *schema.StructValidator, konfig *koanf.Koanf) {
	configuration.JWTSecret = getSecretValue(SecretNames["JWTSecret"], validator, konfig)
	configuration.Session.Secret = getSecretValue(SecretNames["SessionSecret"], validator, konfig)

	if configuration.DuoAPI != nil {
		configuration.DuoAPI.SecretKey = getSecretValue(SecretNames["DUOSecretKey"], validator, konfig)
	}

	if configuration.Session.Redis != nil {
		configuration.Session.Redis.Password = getSecretValue(SecretNames["RedisPassword"], validator, konfig)

		if configuration.Session.Redis.HighAvailability != nil {
			configuration.Session.Redis.HighAvailability.SentinelPassword =
				getSecretValue(SecretNames["RedisSentinelPassword"], validator, konfig)
		}
	}

	if configuration.AuthenticationBackend.LDAP != nil {
		configuration.AuthenticationBackend.LDAP.Password = getSecretValue(SecretNames["LDAPPassword"], validator, konfig)
	}

	if configuration.Notifier != nil && configuration.Notifier.SMTP != nil {
		configuration.Notifier.SMTP.Password = getSecretValue(SecretNames["SMTPPassword"], validator, konfig)
	}

	if configuration.Storage.MySQL != nil {
		configuration.Storage.MySQL.Password = getSecretValue(SecretNames["MySQLPassword"], validator, konfig)
	}

	if configuration.Storage.PostgreSQL != nil {
		configuration.Storage.PostgreSQL.Password = getSecretValue(SecretNames["PostgreSQLPassword"], validator, konfig)
	}

	if configuration.IdentityProviders.OIDC != nil {
		configuration.IdentityProviders.OIDC.HMACSecret = getSecretValue(SecretNames["OpenIDConnectHMACSecret"], validator, konfig)
		configuration.IdentityProviders.OIDC.IssuerPrivateKey = getSecretValue(SecretNames["OpenIDConnectIssuerPrivateKey"], validator, konfig)
	}
}

func getSecretValue(name string, validator *schema.StructValidator, konfig *koanf.Koanf) string {
	configValue := konfig.String(name)
	fileEnvValue := konfig.String(SecretNameToEnvName(name))

	// Error Checking.
	if fileEnvValue != "" && configValue != "" {
		validator.Push(fmt.Errorf("error loading secret (%s): it's already defined in the config file", name))
	}

	// Derive Secret.
	if fileEnvValue != "" {
		content, err := ioutil.ReadFile(fileEnvValue)
		if err != nil {
			validator.Push(fmt.Errorf("error loading secret file (%s): %s", name, err))
		} else {
			// TODO: Test this functionality.
			return strings.TrimRight(string(content), "\n")
		}
	}

	return configValue
}
