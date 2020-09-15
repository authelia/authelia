package validator

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateSecrets checks that secrets are either specified by config file/env or by file references.
func ValidateSecrets(configuration *schema.Configuration, validator *schema.StructValidator, viper *viper.Viper) {
	configuration.JWTSecret = getSecretValue("jwt_secret", validator, viper)
	configuration.Session.Secret = getSecretValue("session.secret", validator, viper)

	if configuration.DuoAPI != nil {
		configuration.DuoAPI.SecretKey = getSecretValue("duo_api.secret_key", validator, viper)
	}

	if configuration.Session.Redis != nil {
		configuration.Session.Redis.Password = getSecretValue("session.redis.password", validator, viper)
	}

	if configuration.AuthenticationBackend.Ldap != nil {
		configuration.AuthenticationBackend.Ldap.Password = getSecretValue("authentication_backend.ldap.password", validator, viper)
	}

	if configuration.Notifier != nil && configuration.Notifier.SMTP != nil {
		configuration.Notifier.SMTP.Password = getSecretValue("notifier.smtp.password", validator, viper)
	}

	if configuration.Storage.MySQL != nil {
		configuration.Storage.MySQL.Password = getSecretValue("storage.mysql.password", validator, viper)
	}

	if configuration.Storage.PostgreSQL != nil {
		configuration.Storage.PostgreSQL.Password = getSecretValue("storage.postgres.password", validator, viper)
	}
}

func getSecretValue(name string, validator *schema.StructValidator, viper *viper.Viper) string {
	configValue := viper.GetString(name)
	fileEnvValue := viper.GetString("authelia." + name + ".file")

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
			return strings.ReplaceAll(string(content), "\n", "")
		}
	}

	return configValue
}
