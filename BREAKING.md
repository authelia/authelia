Breaking changes
================

Since Authelia is still under active development, it is subject to breaking changes. It's
recommended not to use the 'latest' Docker image tag blindly but pick a version instead
and read this documentation before upgrading. This is where you will get information about
breaking changes and about what you should do to overcome those changes.

## Breaking in v4.21.0
* New LDAP attribute `display_name_attribute` has been introduced, defaults to value: `displayname`.
* New key `displayname` has been introduced into the file based user database.

These are utilised to greet the logged in user.

If utilising a file based user backend:
* Administrators will need to update users and include the `displayname` key.

**Before:**
```yaml
users:
  john:
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
```
**After:**
```yaml
users:
  john:
    displayname: "John Doe"
    password: "$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
```   
* Users with long-lived sessions will need to recreate the session (logout and login) to propagate the changes.   

## Breaking in v4.20.0
* Authelia's Docker volumes have been refactored. All data should reside within a single volume of `/config`.
All examples have been updated to reflect this change. The entrypoint for the container changed from
`authelia --config /etc/authelia/configuration.yml` to `authelia --config /config/configuration.yml`.

Users migrating to v4.20.0 have two options:
1. Change your container mappings to point to `/config` also change any associated paths in your `configuration.yml` to
represent the new `/config` mappings.
2. Change your container entry point back to `authelia --config /etc/authelia/configuration.yml`
    * **Docker Compose:** `command: authelia --config /etc/authelia/configuration.yml`
    * **Docker Run:** `docker run -d -v /path/on/host:/etc/authelia authelia/authelia:latest authelia --config /etc/authelia/configuration.yml`
    
The team recommends option 1 to unify/simplify troubleshooting for support related issues.

## Breaking in v4.18.0
* Secrets stored directly in ENV are now removed from Authelia. They have been replaced with file
secrets. If you still have not moved feel free to contact the team for assistance, otherwise the
[documentation](https://docs.authelia.com/configuration/secrets.html) has instructions on how to utilize these.

## Breaking in v4.15.0
* Previously if a configuration value did not exist we ignored it. Now we will error if someone has
specified either an unknown configuration key or one that has changed. In the instance of a changed
key a more specific error is intended. This may cause some people who have not updated their config
to see new errors.
* Authelia now checks the Notifier is configured correctly before becoming available. If the 
SMTP Notifier is used and the configuration is wrong or there is something wrong with the server
Authelia will not start. If the File Notifier is used and the file is not writable Authelia will
not start.
* Authelia v3 migration tools are being removed in this release due to the length of time which
has passed since v4 release. Older versions will still be available for migration if needed.

### Deprecation Notice(s)
* Environment variable secrets are insecure and have been replaced by a file based alternative
instead of having the plain text secret in the environment variables. In version 4.18.0 the old method
will be completely removed. Read more in the [docs](https://docs.authelia.com/configuration/secrets.html).

## Breaking in v4.10.0
* Revert of `users_filter` purpose. This option now represents the complete search filter again, meaning
there is no more automatic filter computation based on username. This gives the most flexibility.
For instance, this allows administrators to choose whether they want the users to be able to sign in with
username or email.

## Breaking in v4.7.0
* `logs_level` configuration key has been renamed to `log_level`.
* `users_filter` was a search pattern for a given user with the `{0}` matcher replaced with the
actual username. In v4.7.0, `username_attribute` has been introduced. Consequently, the computed
user filter utilised by the LDAP search query is a combination of filters based on the
`username_attribute` and `users_filter`. `users_filter` now reduces the scope of users targeted by
the LDAP search query. For instance if `username_attribute` is set to `uid` and `users_filter` is
set to `(objectClass=person)` then the computed filter is `(&(uid=john)(objectClass=person))`.

## Breaking in v4.0.0
Authelia has been rewritten in Go for better code maintainability and for performance and
security reasons.

The principles stay the same, Authelia is still an authenticating and authorizing proxy.
Some major changes have been made though so that the system is more reliable overall. This
induced breaking the previous data model and the configuration to bring new features but
fortunately migration tools are provided to ease the task.

### Major updates
* The configuration mostly remained the same, only one major key has been added: `jwt_secret`
and one key removed: `secure` from the SMTP notifier as the Go SMTP library default to TLS
if available.
* The Hash router has been removed and replaced with a Browser router. This means that the weird characters
/%23/ and /#/ in the redirection URL can now be safely removed.
* The local storage used for dev purpose was a `nedb` database which was implementing the
same interface as mongo but was not really standard. It has been replaced by a good old
sqlite3 database.
* The model of the database is not compatible with v3. This has been decided to better fit
with Golang libraries.
* Some features have been upgraded such as U2F in order to use the latest security features
available like allowing device cloning detection.
* Furthermore, a top-notch web server implementation (fasthttp) has been selected to allow a
large performance gain in order to use Authelia in demanding environments.

### Data migration tools
An authelia-scripts command is provided to perform the data model migration from a local database
or a mongo database created by Authelia v3 into a target SQL database (sqlite3, mysql, postgres)
supported by Authelia v4.

Example of usage:
```
# Migrate a local database into the targeted database defined in config-v4.yml with Docker
docker run --rm -v /path/to/config-v4.yml:/config.yml -v /old/db/path:/db authelia/authelia:4.14.2 authelia migrate local --config=/config.yml --db-path=/db
    
# Migrate a mongo database into the targeted database defined in config-v4.yml with Docker
docker run --rm -v /path/to/config-v4.yml:/config.yml authelia/authelia:4.14.2 authelia migrate mongo --config=/config.yml --url=mongodb://myuser:mypassword@mymongo:27017 --database=authelia
```

Those commands migrate TOTP secrets, U2F devices, authentication traces and user preferences so
that the migration is almost seamless for your users.

The identity verification tokens are not migrated though since their format has changed. However they were
made to expire after a few minutes anyway. Consequently, the users who initiated a device registration process
which has not been completed before the migration will have to restart the device registration process for their
device. This is because their identity verification token will not be usable in v4.

## Breaking in v3.14.0

### Headers in nginx configuration
In order to support Traefik as a third party proxy interacting with Authelia some changes had to be made
to Authelia and the nginx proxy configuration.

The `Host` header is not used anymore by Authelia in any way. It was previously used to compute the url of the link that is
sent by Authelia for confirming the identity of the user. In the new version X-Forwarded-Proto, X-Forwarded-Host
headers are used to build the URL.

Authelia endpoint /api/verify does not produce the `Redirect` header containing the target URL the user is trying to visit.
This header was used in early versions to redirect the user to the login portal providing the target URL as a query parameter.
However this target URL can be computed automatically with the following statement:
```
set                         $target_url $scheme://$http_host$request_uri;
```

## Breaking in v3.11.0

### ACL configuration
ACL definition in the configuration file has been updated to allow more authorization use cases.
The change basically removed the three categories "any", "groups" and "users" to introduce an
iptables-like format where the authorization policy is just an ordered list of rules with a few
attributes among which the attribute called `subject` used to map old categories.

So in order to upgrade from prior version, you simply need to flatten the rules you already have and
use the `subject` attribute to map your rules from the previous categories into the list. For `any`
rules, just don't specify the subject attribute, this rule will then apply to any user. For group-based
rules you can use `subject: 'group:mygroup'` where `mygroup` is the group you set authorizations for.
For user-based rules, use `subject: 'user:myuser'` where `myuser` is the user you set authorizations for.

Please note that in the new system, the first matching rule applies and the next ones are not taken into
account. If no rule apply, the default policy still applies and if no default policy is provided, the `deny`
policy applies.
