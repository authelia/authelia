Breaking changes
================

Since Authelia is still under active development, it is subject to breaking changes. We then recommend you don't blindly use the latest
Docker image but pick a version instead and check this file before upgrading. This is where you will get information about breaking changes and about what you should do to overcome those changes.

## Breaking in v4.0.0

Authelia has been rewritten in Go for better performance and reliability.

### Model of U2F devices in MongoDB

The model of U2F devices stored in MongoDB has been updated to better fit with the Go library handling U2F keys.

### Removal of flag secure for SMTP notifier

The go library for sending e-mails automatically switch to TLS if possible according to https://golang.org/pkg/net/smtp/#SendMail.

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

    set                         $target_url $scheme://$http_host$request_uri;


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