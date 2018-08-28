Release Notes - Version 3.9.2
-----------------------------
* Put back link to Gitter instead of Slack.

Release Notes - Version 3.9.1
-----------------------------
* Split the README in several parts.
* Fix Kubernetes configuration file for Authelia.

Release Notes - Version 3.9.0
-----------------------------
Features:
* Add support for file users database to replace LDAP in development
environments.
* Add authentication configuration options for mongo and redis.

Configuration changes:
* [BREAKING] `ldap` key has been nested in `authentication_backend`.
* New `username` and `password` options for mongo storage.
* New `password` option for redis.

Release Notes - Version 3.8.3
-----------------------------
* Fix ECONNRESET issues when LDAP queries failed. (#261)

Release Notes - Version 3.8.2
-----------------------------
* Fix publication to NPM.

Release Notes - Version 3.8.1
-----------------------------
* Fix publication to NPM.

Release Notes - Version 3.8.0
-----------------------------
Features:
* Add support for Kubernetes nginx ingress controller.
* Add example configuration for kubernetes.
* Disable forms when authentication is in progress.
* Make most of configuration options optional and create a minimal configuration.
* Introduce helmet package to improve security.

Configuration changes:
* [Breaking] `redirect=` in nginx configuration has been replaced by `rd=` to be
be compatible with Kubernetes ingress controller.

Release Notes - Version 3.7.1
-----------------------------
Configuration change:
* storage.mongo now contains two keys: `url` and `database`.

Release Notes - Version 3.7.0
-----------------------------
Features:
* Support basic authorization for single factor endpoints.
* Add issuer and label in TOTP otp url.
* Improve UI of second factor page.
* Use SHA512 password encryption algorithm of LDAP.
* Improve security of Authelia website.
* Support for default redirection url.
* Support for session inactivity timeout.

Bugs:
* Fix U2F factor not working in Firefox

