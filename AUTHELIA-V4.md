# Authelia v4

Authelia has been rewritten in Go for better code maintainability and for performance and security reasons.

The principles stay the same, Authelia is still an authenticating and authorizing proxy. Some major changes have been made though so
that the system is more reliable overall.

Majors changes:
* The configuration mostly remained the same, only one major key has been added: `jwt_secret` and one key removed: `secure` from the
SMTP notifier as the Go SMTP library default to TLS if available.
* The local storage previously used as a replacement of mongo for dev purpose was a `nedb` database which was implementing the same interface
as mongo but was not really standard. It has been replaced by a good old sqlite3 database.
* The model of the database is not compatible with v3. This has been decided to better fit with Golang libraries.
* Some features have been upgraded such as U2F in order to use the latest security features available like allowing device cloning detection.
* Furthermore, a top-notch web server implementation (fasthttp) has been selected to allow a large performance gain in order to use Authelia in demanding environments.


## Migration from v3 to v4

Please note that the migration is breaking the configuration and the data model. Therefore the actions proposed (as of now) to do the migration will make you lose previously registered devices that you'll need to register again in v4.

### Automatic Steps

Since v4 is in beta phase, manual steps are provided for those who are ready to lose their configuration or bootstrap a new instance.
However a migration script will be provided later on. Help for writing this script will be welcome by the way.

### Manual Steps

* Add the `jwt_secret` key in the configuration along with the value of the secret. This secret is used to generate expirable JWT tokens
for operations requiring identity validation.
* Remove the `secure` key of your SMTP notifier configuration as the Go implementation of the SMTP library uses TLS by default if available.

#### If using the local storage
* Remove the directory of the storage (beware you will lose your previous configuration: U2F, TOTP devices). Replace the path with a path to a sqlite3 database,
it is the new standard way of storing data in Authelia.

#### If using the mongo storage
* Flush your collections (beware you will lose your previous configuration: U2F, TOTP devices). New collections will be created by Authelia.
