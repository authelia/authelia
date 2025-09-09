---
title: "Measures"
description: "An overview of the security measures Authelia implements."
summary: "An overview of the security measures Authelia implements."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 420
toc: true
aliases:
  - /docs/security/measures.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Protections against return oriented programming attacks and general hardening

Authelia is built as a position independent executable which makes Return Oriented Programming (ROP) attacks
significantly more difficult to execute reliably.

In addition, it is built as a dynamically linked binary with full relocation read-only support, making this and several
other traditional binary weaknesses significantly more difficult to exploit.

## Protections against unnecessary attack surface

Authelia takes a proactive approach to security and as such has spent countless hours reducing the attack surface by
removing unnecessary components from our architecture.

Specifically we've developed our own docker base container. This base container is specifically designed to have a low
attack surface only having the minimum binaries required to support the container. You can verify the benefits of this
proactive measure by using tools like [trivy](https://trivy.dev/latest/) on various containers. An example running it
on Authelia is shown below:

```shell
docker pull docker.io/authelia/authelia:latest && \
trivy image docker.io/authelia/authelia:latest && \
docker run docker.io/authelia/authelia:latest authelia build-info
```

Example output:

```text
Report Summary

┌───────────────────────────────────────────────────┬──────────┬─────────────────┬─────────┐
│                      Target                       │   Type   │ Vulnerabilities │ Secrets │
├───────────────────────────────────────────────────┼──────────┼─────────────────┼─────────┤
│ docker.io/authelia/authelia:latest (ubuntu 24.04) │  ubuntu  │        0        │    -    │
├───────────────────────────────────────────────────┼──────────┼─────────────────┼─────────┤
│ app/authelia                                      │ gobinary │        0        │    -    │
└───────────────────────────────────────────────────┴──────────┴─────────────────┴─────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)

Last Tag: v4.39.8
State: tagged clean
Branch: v4.39.8
Commit: 5d90442e07cc695c61036ac1a539c0b942ebc71d
Build Number: 48388
Build OS: linux
Build Arch: amd64
Build Compiler: gc
Build Date: Tue, 02 Sep 2025 10:42:14 +1000
Development: false
Extra:

Go:
    Version: go1.25.0 X:nosynchashtriemap
    Module Path: github.com/authelia/authelia/v4
    Executable Path: github.com/authelia/authelia/v4/cmd/authelia
```

## Protection against cookie theft

Authelia sets several
[cookie attributes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#attributes) to help prevent
cookie theft:

1. `HttpOnly` is set, forbidding client-side code like javascript from access to the cookie.
2. `Secure` is set, forbidding the browser from sending the cookie to sites which do not use the [HTTPS] scheme.
3. `SameSite` is set to `Lax`, which prevents it being sent over cross-origin requests. An option to adjust this value
    exists but is not recommended.

Read about these attributes in detail on the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie).

## Protection against multi-domain cookie attacks

Since Authelia uses multi-domain cookies to perform single sign-on, an attacker who poisoned a user's DNS cache can
easily retrieve the user's cookies by making the user send a request to one of the attacker's IPs.

This is technically mitigated by the `Secure` attribute set in cookies by Authelia, however it's still advisable to
only use [HTTPS] connections with valid certificates and enforce it with HTTP Strict Transport Security ([HSTS]) which
will prevent domains from serving over [HTTP] at all as long as the user has visited the domain before. This means even
if the attacker poisons DNS they are unable to get modern browsers to connect to a compromised host unless they can also
obtain the certificate.

Note that using [HSTS] has consequences, and you should do adequate research into understanding [HSTS] before you enable
it. For example the [nginx blog] has a good article helping users understand it.

## Protection against username enumeration and password brute-force attacks

Authelia adaptively delays authentication attempts based on the mean (average) of the previous 10 successful attempts
in addition to a small random interval of time. The result of this delay is that it makes it incredibly difficult to
determine if the unsuccessful login was the result of a bad password, a bad username, or both. The random interval of
time is anything between 0 milliseconds and 85 milliseconds.

When Authelia first starts it assumes the last 10 attempts took 1000 milliseconds each. As users login successfully it
quickly adjusts to the actual time the login attempts take. This process is independent of the login backend you have
configured.

The cost of this is low since in the instance of a user not existing it just stops processing the request to delay the
login. Lastly the absolute minimum time authentication can take is 250 milliseconds. Both of these measures also have
the added effect of creating an additional delay for all authentication attempts increasing the time that a brute-force
attack will take, this combined with regulation greatly delays brute-force attacks and the effectiveness of them in
general.

## Protections against password brute-force attacks

Authelia implements a variety of measures to prevent an attacker brute-forcing passwords if they somehow obtain the file
used by the file authentication provider.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The LDAP authentication provider honors the password modify extended operation if available which delegates this task to
the LDAP server.
{{< /callout >}}

First and foremost Authelia only uses very secure hashing algorithms with sane and secure defaults. The first and
default hashing algorithm we use is Argon2id which is currently considered the most secure hashing algorithm. We also
support SHA512, which previously was the default.

Secondly Authelia uses salting with all hashing algorithms. These salts are generated with a random string generator,
which is seeded every time it's used by a cryptographically secure 1024bit prime number. This ensures that even if an
attacker obtains the file, each password has to be brute forced individually.

Lastly Authelia's implementation of Argon2id is highly tunable. You can tune the key length, salt used, iterations
(time), parallelism, and memory usage. To read more about this, please read how to
[configure](../../configuration/first-factor/file.md) file authentication.

## Protection against request brute-force attacks

Authelia implements a tokenized bucket rate limiter on specific endpoints which greatly reduces the chances these
endpoints can be brute-forced for various outcomes including guessing secret values, or inundating an inbox with emails.

These rate limiters are applied on a per-IP basis and can be
[configured](../../configuration/miscellaneous/server-endpoint-rate-limits.md) depending on a particular use case.

## User profile and group membership always kept up-to-date (LDAP authentication provider)

This measure is unrelated to the File authentication provider.

Authelia by default refreshes the user's profile and membership every 5 minutes. This ensures that if you alter a users
groups in LDAP that their new groups are obtained relatively quickly in order to adjust their access level for
applications secured by Authelia.

Additionally, it will invalidate any session where the user could not be retrieved from LDAP based on the user filter,
for example if they were deleted or disabled provided the user filter is set correctly. These updates occur when a user
accesses a resource protected by Authelia. This means you should ensure disabled users or users with expired passwords
are not obtainable using the LDAP filter, the default filter for Active Directory and several other implementation
defaults implement this behavior. LDAP implementations vary, so please ask if you need some assistance in configuring
this.

These protections can be [tuned](../../configuration/first-factor/ldap.md#refresh-interval) according to your security
policy by changing refresh_interval, however we believe that 5 minutes is a fairly safe interval.

## Storage security measures

We force users to encrypt vulnerable data stored in the database. It is strongly advised you do not give this encryption
key to anyone. In the instance of a database installation that multiple users have access to, you should aim to ensure
that users who have access to the database do not also have access to this key.

The encrypted data in the database is as follows:

|               Table               |    Column    |                                                Rational                                                |
|:---------------------------------:|:------------:|:------------------------------------------------------------------------------------------------------:|
|        totp_configurations        |    secret    | Prevents a [Leaked Database](#leaked-database) or [Bad Actors](#bad-actors) from compromising security |
|         webauthn_devices          |  public_key  |                     Prevents [Bad Actors](#bad-actors) from compromising security                      |
| oauth2_authorization_code_session | session_data |                     Prevents [Bad Actors](#bad-actors) from compromising security                      |
|    oauth2_access_token_session    | session_data |                     Prevents [Bad Actors](#bad-actors) from compromising security                      |
|   oauth2_refresh_token_session    | session_data |                     Prevents [Bad Actors](#bad-actors) from compromising security                      |
|    oauth2_pkce_request_session    | session_data |                     Prevents [Bad Actors](#bad-actors) from compromising security                      |
|   oauth2_openid_connect_session   | session_data |                     Prevents [Bad Actors](#bad-actors) from compromising security                      |

### Leaked Database

A leaked database can reasonably compromise security if there are credentials that are not encrypted. Columns encrypted
for this purpose prevent this attack vector.

### Bad Actors

A bad actor who has the SQL password and access to the database can theoretically change another users credential, this
theoretically bypasses authentication. Columns encrypted for this purpose prevent this attack vector.

A bad actor may also be able to use data in the database to bypass 2FA silently depending on the credentials. In the
instance of the U2F public key this is not possible, they can only change it which would eventually alert the user in
question. But in the case of TOTP they can use the secret to authenticate without knowledge of the user in question.

### Encryption key management

You must supply the encryption key in the recommended method of a [secret](../../configuration/methods/secrets.md) or in
one of the other [methods available for configuration](../../configuration/methods/introduction.md).

If you wish to change your encryption key for any reason you can do so using the following steps:

1. Run the `authelia --version` command to determine the version of Authelia you're running and either download that
   version or run another container of that version interactively. All the subsequent commands assume you're running
   the `authelia` binary in the current working directory. You will have to adjust this according to how you're running
   it.
2. Run the `./authelia storage encryption change-key --help` command.
3. Stop Authelia.
   * You can skip this step, however note that any data changed between the time you make the change and the time when
   you stop Authelia i.e. via user registering a device; will be encrypted with the incorrect key.
4. Run the `./authelia storage encryption change-key` command with the appropriate parameters.
   * The help from step 1 will be useful here. The easiest method to accomplish this is with the `--config`,
   `--encryption-key`, and `--new-encryption-key` parameters.
5. Update the encryption key Authelia uses on startup.
6. Start Authelia.

## Notifier security measures (SMTP)

The SMTP Notifier implementation does not allow connections that are not secure without changing default configuration
values.

As such all SMTP connections require the following:

1. A TLS Connection (StartTLS or implicit) has been negotiated before authentication or sending emails (_unauthenticated
 connections require it as well_)
2. Valid X509 Certificate presented to the client during the TLS handshake

There is an option to disable both of these security measures however they are __not recommended__.

The following configuration options exist to configure the security level in order of most preferable to least
preferable:

### Configuration Option: certificates_directory

You can configure a [certificates_directory] option which contains certificates for Authelia to trust. These certificates
can either be CA's or individual public certificates that should be trusted. These are added in addition to the
environments PKI trusted certificates if available. This is useful for trusting a certificate that is self-signed without
drastically reducing security. This is the most recommended workaround to not having a valid PKI trusted certificate as
it gives you complete control over which ones are trusted without disabling critically needed validation of the identity
of the target service.

Read more in the [certificates_directory] documentation for this option.

[certificates_directory]: ../../configuration/miscellaneous/introduction.md#certificates_directory
[certificates directory]: #configuration-option-certificates_directory

### Configuration Option: tls.skip_verify

The [tls.skip_verify](../../configuration/notifications/smtp.md#tls) option allows you to skip verifying the certificate
entirely which is why [certificates directory] is preferred over this. This will effectively mean you cannot be sure the
certificate is valid which means an attacker via DNS poisoning or MITM attacks could intercept emails from Authelia
compromising a user's security without their knowledge.

### Configuration Option: disable_require_tls

Authelia by default ensures that the SMTP server connection is secured via TLS prior to sending sensitive information.

The [disable_require_tls](../../configuration/notifications/smtp.md#disable_require_tls) option disables this
requirement which means the emails may be sent in cleartext. This is the least secure option as it effectively removes
the validation of SMTP certificates and makes using an encrypted connection with TLS optional.

This means not only can the vulnerabilities of the [skip_verify](#configuration-option-tlsskip_verify) option be
exploited, but any router or switch along the route of the email which receives the packets could be used to silently
exploit the cleartext nature of the connection to manipulate the email in transit.

This is only usable currently with authentication disabled (_comment out the password_), and as such is only an option
for SMTP servers that allow unauthenticated relaying (bad practice).

### SMTP Ports

All SMTP connections begin as [cleartext], and then negotiate to upgrade to a secure TLS connection via StartTLS.

The [`submissions` service][service-submissions] (_typically port 465_) is an exception to this rule, where the
connection begins immediately secured with TLS (_similar to HTTPS_). When the configured [scheme for
SMTP][docs-config-smtp-port] is set to `submissions`, Authelia will initiate TLS connections without requiring StartTLS
negotiation.

When the `submissions` service port is available, it [should be preferred][port-465] over any StartTLS port for
submitting mail.

**NOTE:** Prior to 2018, port 465 was previously assigned for a similar purpose known as [`smtps`][port-465] (_A TLS
only equivalent of the `smtp` port 25_), which it had been deprecated for. Port 465 has since been re-assigned for only
supporting mail submission (_which unlike SMTP transfers via port 25, [requires authentication][smtp-auth]_), similar
to port 587 (_the `submission` port, a common alternative that uses StartTLS instead_).

[docs-config-smtp-port]: ../../configuration/notifications/smtp.md#address
[cleartext]: https://cwe.mitre.org/data/definitions/312.html
[service-submissions]: https://datatracker.ietf.org/doc/html/rfc8314#section-7.3
[port-465]: https://datatracker.ietf.org/doc/html/rfc8314#section-3.3
[smtp-auth]: https://datatracker.ietf.org/doc/html/rfc6409#section-4.3

## Protection against open redirects

Authelia protects your users against open redirect attacks by always checking if redirection URLs are pointing
to a subdomain of the domain protected by Authelia. This prevents phishing campaigns tricking users into visiting
infected websites leveraging legit links.

## Mutual TLS

For the best security protection, configuration with TLS is highly recommended. TLS is used to secure the connection
between the proxies and Authelia instances meaning that an attacker on the network cannot perform a man-in-the-middle
attack on those connections. However, an attacker on the network can still impersonate proxies but this can be prevented
by configuring mutual TLS.

Mutual TLS brings mutual authentication between Authelia and the proxies. Any other party attempting to contact Authelia
would not even be able to create a TCP connection. This measure is recommended in all cases except if you already
configured some kind of ACLs specifically allowing the communication between proxies and Authelia instances like in a
service mesh or some kind of network overlay.

To configure mutual TLS, please refer to [this document](../../configuration/miscellaneous/server.md#client_certificates)

## Additional security

### Reset Password

It's possible to disable the reset password functionality and is an optional adjustment to consider for anyone wanting
to increase security. See the [configuration](../../configuration/first-factor/introduction.md#disable)
for more information.

### Session security

We have a few options to configure the security of a session. The main and most important one is the session secret.
This is used to encrypt the session data when it is stored in the [Redis](../../configuration/session/redis.md) key value
database. The value of this option should be long and as random as possible. See more in the
[documentation](../../configuration/session/introduction.md#secret) for this option.

The validity period of session is highly configurable. For example, in a highly security conscious domain you could
set the session [remember_me](../../configuration/session/introduction.md#remember_me) to 0 to disable this
feature, and set the [expiration](../../configuration/session/introduction.md#expiration) to 2 hours and the
[inactivity](../../configuration/session/introduction.md#inactivity) of 10 minutes. Configuring the session security in this
manner would mean if the cookie age was more than 2 hours or if the user was inactive for more than 10 minutes the
session would be destroyed.

### Response Headers

This document previously detailed additional per-proxy configuration options that could be utilized in a proxy to
improve security. These headers are now documented here and implemented by default in all responses due to the fact
the experience should be the same regardless of which proxy you're utilizing and the area is rapidly evolving.

Users who need custom behaviors in this area can submit a request or remove/replace the headers as necessary. These
headers will evolve over time just as the web standards and security recommendations evolve. These headers prevent
loading Authelia in specific scenarios, primarily in an
[Inline Frame](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/iframe) which is generally considered a high
security risk.

The [OWASP](https://owasp.org/) helpful
[HTTP Security Response Headers Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html)
was used as a basis for most of the decisions regarding these headers. Users who which to customize the behavior should
consider this cheat sheet mandatory reading before they do so.

#### X-XSS-Protection

__Value:__ N/A
__Endpoints:__ All
__Status:__ Unsupported Non-standard

We do not include this header as this feature is not present in any modern browser and could introduce vulnerabilities
if enabled at all. Going forward [CORS], [CORP], CORB, and [COEP] are the standards for browser-centric site security.
See the [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection) for more information.

#### X-Content-Type-Options

__Value:__ `nosniff`
__Endpoints:__ All
__Status:__ Supported Standard

Prevents MIME type sniffing. See the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options) for more information.

#### Referrer-Policy

__Value:__ `strict-origin-when-cross-origin`
__Endpoints:__ All
__Status:__ Supported Standard

Sends only the origin as the referrer in cross-origin requests, but sends the origin, path, and query string in
same-origin requests. See the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy) for more information.

#### X-Frame-Options

__Value:__ `DENY`
__Endpoints:__ All
__Status:__ Supported Standard

Prevents Authelia rendering in a `frame`, `iframe`, `embed`, or `object` element. See the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options) for more information.

#### Permissions-Policy

__Value:__ `accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=(), interest-cohort=()`
__Endpoints:__ All
__Status:__ Supported Standard

Disables browser features not required by Authelia including the
[Federated Learning of Cohorts](https://en.wikipedia.org/wiki/Federated_Learning_of_Cohorts). It should be noted while
this is a supported standard individual features of the permissions policy may not be supported by some browsers or
browser configurations. See the [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Permissions-Policy) and
the [Permissions Policy website](https://www.permissionspolicy.com/) for
more information.

#### X-DNS-Prefetch-Control

__Value:__ `off`
__Endpoints:__ All
__Status:__ Non-standard

Prevents browsers from performing a DNS prefetch for links displayed on Authelia pages. Not all browsers support this.
See the [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-DNS-Prefetch-Control) for more information.

#### Pragma

__Value:__ `no-cache`
__Endpoints:__ API
__Status:__ Supported Standard

Disables caching of API requests on HTTP/1.0 browsers. See the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Pragma) for more information.

#### Cache-Control

__Value:__ `no-store`
__Endpoints:__ API
__Status:__ Supported Standard

Disables caching responses entirely on HTTP/1.1 browsers. See the
[MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control) for more information.

### More protections measures with fail2ban

If you are running fail2ban, adding a filter and jail for Authelia can reduce load on the application / web server.
Fail2ban will ban IPs exceeding a threshold of repeated failed logins at the firewall level of your host. If you are
using Docker, the Authelia log file location has to be mounted from the host system to the container for
fail2ban to access it.

Create a configuration file in the `filter.d` folder with the content below. In Debian-based systems the folder is
typically located at `/etc/fail2ban/filter.d`.

```ini
# Fail2Ban filter for Authelia

# Make sure that the HTTP header "X-Forwarded-For" received by Authelia's backend
# only contains a single IP address (the one from the end-user), and not the proxy chain
# (it is misleading: usually, this is the purpose of this header).

# the failregex rule counts every failed 1FA attempt (first line, wrong username or password) and failed 2FA attempt
# second line) as a failure.
# the ignoreregex rule ignores info and warning messages as all authentication failures are flagged as errors
# the third line catches incorrect usernames entered at the password reset form
# the fourth line catches attempts to spam via the password reset form or 2fa device reset form. This requires debug logging to be enabled

[Definition]
failregex = ^.*Unsuccessful (1FA|TOTP|Duo|U2F) authentication attempt by user .*remote_ip"?(:|=)"?<HOST>"?.*$
            ^.*user not found.*path=/api/reset-password/identity/start remote_ip"?(:|=)"?<HOST>"?.*$
            ^.*Sending an email to user.*path=/api/.*/start remote_ip"?(:|=)"?<HOST>"?.*$

ignoreregex = ^.*level"?(:|=)"?info.*
              ^.*level"?(:|=)"?warning.*
```

Modify the `jail.local` file. In Debian-based systems the folder is typically located at `/etc/fail2ban/`. If the file
does not exist, create it by copying the jail.conf `cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local`. Add an
Authelia entry to the "Jails" section of the file:

```ini
[authelia]
enabled = true
port = http,https,9091
filter = authelia
logpath = /path-to-your-authelia.log
maxretry = 3
bantime = 1d
findtime = 1d
chain = DOCKER-USER
action = iptables-allports[name=authelia]
```

If you are not using Docker remove the line "chain = DOCKER-USER". You will need to restart the fail2ban service for the
changes to take effect.

## Container privilege de-escalation

Authelia will run as the root user and group by default, there are two options available to run as a non-root user and
group.

It is recommended, which ever approach you take, in order to secure the sensitive files Authelia requires access to, that you
make sure that the mode (chmod) of the files does not inadvertently allow read access to the files by users who do not need access
to them.

Examples:

If you wanted to run Authelia as UID 8000, and wanted the GID of 9000 to also have read access to the files
you might do the following assuming the files were in the relative path `.data/authelia`:

```shell
chown -r 8000:9000 .data/authelia
find .data/authelia/ -type d -exec chmod 750 {} \;
find .data/authelia/ -type f -exec chmod 640 {} \;
```

If you wanted to run Authelia as UID 8000, and wanted the GID of 9000 to also have write access to the files
you might do the following assuming the files were in the relative path `.data/authelia`:

```shell
chown -r 8000:9000 .data/authelia
find .data/authelia/ -type d -exec chmod 770 {} \;
find .data/authelia/ -type f -exec chmod 660 {} \;
```

### Docker user directive

The docker user directive allows you to configure the user the entrypoint runs as. This is generally the most secure
option for containers as no process accessible to the container ever runs as root which prevents a compromised container
from exploiting unnecessary privileges.

The directive can either be applied in your `docker run` command using the `--user` argument or by
the docker compose `user:` key. The examples below assume you'd like to run the container as UID 8000 and GID 9000.

Example for the docker CLI:

```shell
docker run --user 8000:9000 -v /authelia:/config authelia/authelia:latest
```

Example for docker compose:

```yaml {title="compose.yml"}
services:
  authelia:
    image: authelia/authelia
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    user: 8000:9000
    volumes:
      - ./authelia:/config
```

Running the container in this way requires that you manually adjust the file owner at the very least as described above.
If you do not do so it will likely cause Authelia to exit immediately. This option takes precedence over the PUID and
PGID environment variables below, so if you use it then changing the PUID and PGID have zero effect.

### PUID/PGID environment variables using the entrypoint

The second option is to use the `PUID` and `PGID` environment variables. When the container entrypoint is executed
as root, the entrypoint automatically runs the Authelia process as this user. An added benefit of using the environment
variables is the mounted volumes ownership will automatically be changed for you. It is still recommended that
you run the find chmod examples above in order to secure the files even further especially on servers multiple people
have access to.

The examples below assume you'd like to run the container as UID 8000 and GID 9000.

Example for the docker CLI:

```shell
docker run -e PUID=8000 -e PGID=9000 -v /authelia:/config authelia/authelia:latest
```

Example for docker compose:

```yaml {title="compose.yml"}
services:
  authelia:
    image: authelia/authelia
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    environment:
      PUID: 8000
      PGID: 9000
    volumes:
      - ./authelia:/config
```

## Privacy

### Opaque Identifiers

Where possible we utilize opaque identifiers which link to user accounts. The primary example at this time is
[OpenID Connect](../../integration/openid-connect/introduction.md) which utilizes an opaque identifier for the
[subject claim](../../integration/openid-connect/introduction.md#openid).

This will also be utilized in the future for the [WebAuthn](../authentication/security-key) passwordless flow
(discoverable logins).

[HTTP]: https://developer.mozilla.org/en-US/docs/Glossary/http
[HTTPS]: https://developer.mozilla.org/en-US/docs/Glossary/https
[HSTS]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
[nginx blog]: https://www.nginx.com/blog/http-strict-transport-security-hsts-and-nginx/
[CORS]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
[CORP]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Cross-Origin_Resource_Policy_(CORP)
[COEP]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cross-Origin-Embedder-Policy
