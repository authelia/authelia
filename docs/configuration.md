# Configuration

Authelia is highly configurable thanks to a configuration file. 
There is a documented template configuration, called
[config.template.yml](../config.template.yml), at the root of the
repository. All the details are documented there.

When running **Authelia**, you can specify your configuration file by passing
the file path as the first argument of **Authelia**.

    $ authelia --config config.custom.yml


## Secrets

Configuration of Authelia requires some secrets or passwords. Please
note that the recommended way to set secrets in Authelia is to use
environment variables.

A secret in Authelia configuration could be set by providing the
environment variable prefixed by AUTHELIA_ and with name equals to
the capitalized path of the configuration key and with dots replaced
by underscores.

For instance the LDAP password is identified by the path
**authentication_backend.ldap.password**, so this password could
alternatively be set using the environment variable called
**AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD**.

If for some reason you prefer keeping the secrets in the configuration
file, be sure to apply the right permissions to the file in order to
prevent secret leaks if an another application gets compromised on your
server. The UNIX permissions should probably be something like 600.