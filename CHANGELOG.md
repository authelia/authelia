Release Notes - Version 3.15.0
------------------------------
* Change license from MIT to Apache 2.0.

Release Notes - Version 3.14.0
------------------------------
* [BREAKING] Add official support for Traefik with a dedicated suite.
* Add support for network-based ACL rules allowing to apply different authorization strategies on different networks.
* Several bug fixes (unusual error message when using U2F, X-Forwarded-User and X-Forwarded-Groups was not propagated on bypassed endpoints).

Release Notes - Version 3.13.0
------------------------------
* Rewrite Authelia portal in Typescript.
* Intoduce concept of suites and authelia-scripts.
* Add official support for Kubernetes and a suite.
* Improve documentation for nginx.
* Fix bypass policy not properly handled.
* Implement Duo push notification as a second factor.
* Display only available 2FA options (U2F if supported in browser, Duo push if configured).

Release Notes - Version 3.12.0
------------------------------
* Add logs to troubleshoot LDAP sanitizer.
* Add {uid} placeholder for LDAP search queries for groups.

Release Notes - Version 3.11.0
------------------------------
* [BREAKING] Flatten ACL rules to enable some use cases. Configuration of ACLs
must be updated.
* Fix open redirection threat.
* Define minimum level of authentication required for a resource in ACL to be
authorized.
* Allow Authelia to be built with different themes.
* Fix bug in hash matching when using file-based users database.
* Fix dead link in documentation.

Release Notes - Version 3.10.0
------------------------------
* Add docker-compose for deploying Authelia on Swarm*.
* Add "keep me logged in" checkbox in first factor page.
* Fix U2F compatiblity with Firefox.
* Bump dependencies to fix vulnerabilities reported by snyk.
* Improve documentation for dev setup.

Release Notes - Version 3.9.5
-----------------------------
* Fix images in README in NPM.

Release Notes - Version 3.9.4
-----------------------------
* Update Authelia icon & add documentation image.
* Add snyk badge

Release Notes - Version 3.9.3
-----------------------------
* Fix npm publication.
* Use IP coming from X-Forwarded-For header in logs.
* Fix CONTRIBUTORS.md.

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
* Fix ECONNRESET issues when LDAP queries failed. (#261).

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

