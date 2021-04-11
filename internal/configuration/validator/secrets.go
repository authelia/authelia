package validator

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// SecretNameToEnvName converts a secret name into the env name.
func SecretNameToEnvName(secretName string) (envName string) {
	return "authelia." + secretName + ".file"
}

func isSecretKey(value string) (isSecretKey bool) {
	for _, secretKey := range SecretNames {
		if value == secretKey || value == SecretNameToEnvName(secretKey) {
			return true
		}
	}

	return false
}

// ValidateSecrets checks that secrets are either specified by config file/env or by file references.
func ValidateSecrets(configuration *schema.Configuration, validator *schema.StructValidator, viper *viper.Viper) {
	configuration.JWTSecret = getSecretValue(SecretNames["JWTSecret"], validator, viper)
	configuration.Session.Secret = getSecretValue(SecretNames["SessionSecret"], validator, viper)

	if configuration.DuoAPI != nil {
		configuration.DuoAPI.SecretKey = getSecretValue(SecretNames["DUOSecretKey"], validator, viper)
	}

	if configuration.Session.Redis != nil {
		configuration.Session.Redis.Password = getSecretValue(SecretNames["RedisPassword"], validator, viper)

		if configuration.Session.Redis.HighAvailability != nil {
			configuration.Session.Redis.HighAvailability.SentinelPassword =
				getSecretValue(SecretNames["RedisSentinelPassword"], validator, viper)
		}
	}

	if configuration.AuthenticationBackend.Ldap != nil {
		configuration.AuthenticationBackend.Ldap.Password = getSecretValue(SecretNames["LDAPPassword"], validator, viper)
	}

	if configuration.Notifier != nil && configuration.Notifier.SMTP != nil {
		configuration.Notifier.SMTP.Password = getSecretValue(SecretNames["SMTPPassword"], validator, viper)
	}

	if configuration.Storage.MySQL != nil {
		configuration.Storage.MySQL.Password = getSecretValue(SecretNames["MySQLPassword"], validator, viper)
	}

	if configuration.Storage.PostgreSQL != nil {
		configuration.Storage.PostgreSQL.Password = getSecretValue(SecretNames["PostgreSQLPassword"], validator, viper)
	}

	if configuration.IdentityProviders.OIDC != nil {
		configuration.IdentityProviders.OIDC.HMACSecret = getSecretValue(SecretNames["OpenIDConnectHMACSecret"], validator, viper)
		configuration.IdentityProviders.OIDC.IssuerPrivateKey = getSecretValue(SecretNames["OpenIDConnectIssuerPrivateKey"], validator, viper)
	}
}

func getSecretValue(name string, validator *schema.StructValidator, viper *viper.Viper) string {
	configValue := viper.GetString(name)
	fileEnvValue := viper.GetString(SecretNameToEnvName(name))

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
