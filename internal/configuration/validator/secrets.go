package validator

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"

	"github.com/authelia/authelia/internal/configuration/schema"
)

//nolint:gocyclo // This check is not required the layout of the function takes care of the complexity and it can't easily be simplified.
// ValidateSecrets checks that secrets are either specified by config file/env or by file references
func ValidateSecrets(configuration *schema.Configuration, validator *schema.StructValidator, viper *viper.Viper) {
	jwtSecret, err := checkSecretValue("jwt_secret", viper)
	if err != nil {
		validator.Push(err)
	} else {
		configuration.JWTSecret = jwtSecret
	}

	if configuration.DuoAPI != nil {
		duoAPISecretKey, err := checkSecretValue("duo_api.secret_key", viper)
		if err != nil {
			validator.Push(err)
		} else {
			configuration.DuoAPI.SecretKey = duoAPISecretKey
		}
	}

	sessionSecret, err := checkSecretValue("session.secret", viper)
	if err != nil {
		validator.Push(err)
	} else {
		configuration.Session.Secret = sessionSecret
	}

	if configuration.Session.Redis != nil {
		redisPassword, err := checkSecretValue("session.redis.password", viper)
		if err != nil {
			validator.Push(err)
		} else {
			configuration.Session.Redis.Password = redisPassword
		}
	}

	if configuration.AuthenticationBackend.Ldap != nil {
		ldapPassword, err := checkSecretValue("authentication_backend.ldap.password", viper)
		if err != nil {
			validator.Push(err)
		} else {
			configuration.AuthenticationBackend.Ldap.Password = ldapPassword
		}
	}

	if configuration.Notifier != nil && configuration.Notifier.SMTP != nil {
		smtpPassword, err := checkSecretValue("notifier.smtp.password", viper)
		if err != nil {
			validator.Push(err)
		} else {
			configuration.Notifier.SMTP.Password = smtpPassword
		}
	}

	if configuration.Storage.MySQL != nil {
		mysqlPassword, err := checkSecretValue("storage.mysql.password", viper)
		if err != nil {
			validator.Push(err)
		} else {
			configuration.Storage.MySQL.Password = mysqlPassword
		}
	}

	if configuration.Storage.PostgreSQL != nil {
		postgresPassword, err := checkSecretValue("storage.postgres.password", viper)
		if err != nil {
			validator.Push(err)
		} else {
			configuration.Storage.PostgreSQL.Password = postgresPassword
		}
	}
}

func checkSecretValue(name string, viper *viper.Viper) (string, error) {
	configValue := viper.GetString(name)
	envValue := viper.GetString("authelia." + name)
	fileEnvValue := viper.GetString("authelia." + name + ".file")

	if envValue != "" && fileEnvValue != "" {
		return "", fmt.Errorf("secret is defined in multiple areas: %s", name)
	} else if configValue == "" && (envValue != "" || fileEnvValue != "") {
		if fileEnvValue != "" {
			content, err := ioutil.ReadFile(fileEnvValue)
			// Replace newlines to prevent editor issues. Note this will not work with CRLF just LF.
			return strings.Replace(string(content), "\n", "", -1), err
		} else if envValue != "" {
			return envValue, nil
		}
	} else if envValue != "" || fileEnvValue != "" {
		err := fmt.Errorf("error loading secret (%s): it's already defined in the config file", name)
		return "", err
	} else {
		return configValue, nil
	}
	return "", nil
}
