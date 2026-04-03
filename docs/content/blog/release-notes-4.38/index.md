---
title: "4.38: Release Notes"
description: "Authelia 4.38 release notes."
summary: "Authelia 4.38 has been released and the following is a guide on all the massive changes."
date: 2024-03-14T06:00:14+11:00
draft: false
weight: 50
categories: ["News", "Release Notes"]
tags: ["releases", "release-notes"]
contributors: ["James Elliott"]
pinned: false
homepage: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia [4.38](https://github.com/authelia/authelia/releases/tag/v4.38.0) is released! This version has several
additional features and improvements to existing features. In this blog post we'll discuss the new features and roughly
what it means for users.

Overall this release adds several major roadmap items. It's quite a big release. We expect a few bugs here and there but
nothing major. It's one of our biggest releases to date, so while it's taken a longer time than usual it's for good
reason we think.

## Foreword

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This section is important to read for all users who are upgrading, especially those who are automatically upgrading.
{{< /callout >}}

There are some changes in this release which deprecate older configurations, you will get a warning about these
deprecations as it's likely in version v5.0.0 we'll remove support for them, however if a log message for
a configuration is a warning then it's just a warning, and can fairly safely be ignored for now. These changes should be
backwards compatible, however mistakes happen. If you find a mistake please kindly let us know.

In addition we advise making the adjustments mentioned in this post to your configuration as several new features will
not be available or even possible without making the necessary adjustments.

It's also important to note that a couple of the configuration changes that give you access to additional features
need to be done together. For example if you use new proxy configurations you must also use the new session
configuration.

## Helm Chart

Those of you interested in the Helm Chart for this please be aware the 0.9.0 release of the chart will include this
version however it will be a breaking change as previously warned. I will have to do a few more checks of the chart
before I release to ensure nothing is missed.

Please see the [Pull Request](https://github.com/authelia/chartrepo/pull/215) for more information and feel free to
try it out and provide feedback in the official
[feedback discussion](https://github.com/authelia/chartrepo/discussions/220).

## On This Page

This blog article is rather large so this serves as an index for all of the areas so you can best find a particular item.

- Key Sections:
  - [OpenID Connect 1.0](#openid-connect-10)
  - [Multi-Domain Protection](#multi-domain-protection)
  - [WebAuthn](#webauthn)
  - [Customizable Authorization Endpoints](#customizable-authorization-endpoints)
  - [Configuration](#configuration)
- Important Changes (these will require manual intervention to gain access to new features, but old ones should still
  work):
  - [Multi-Domain Protection](#changes-multiple-domain-protection) which is important for all users.
  - [Customizable Authorization Endpoints](#changes-customizable-authorization-endpoints) which is important for all
    users.
  - [OpenID Connect 1.0: Multiple JSON Web Keys](#changes-multiple-json-web-keys) which is important for users of the
    OpenID Connect 1.0 Provider features.
  - [Client Authentication Method (Token Endpoint)](#client-authentication-method-token-endpoint) may require
    administrators provide configuration after the update.
  - [OpenID Connect 1.0: Other Notable Changes](#other-notable-changes)
- Important Modification Considerations (changes that you are not required to make but may run into issues if you're
  not careful):
  - [Server Listener](#server-listener) changes

## OpenID Connect 1.0

As part of our ongoing effort for comprehensive support for [OpenID Connect 1.0] we'll be introducing several important
features. Please see the [roadmap](../../roadmap/active/openid-connect-1.0-provider.md) for more information.

Those of you familiar with the various specifications are going to notice a few features which are very large steps
towards the Financial-grade API Security Profile and OAuth 2.0 Security Best Current Practice, this is because we are
putting a lot of time into implementing security and privacy first features.

We have also put a lot of effort into aligning configuration names with the relevant specifications to make it easier
in the future to implement features such as Dynamic Client Registration.

#### OAuth 2.0 Client Credentials Flow

This release includes the machine-based Client Credentials Flow which can be used to programmatically obtain tokens, as
well as ways to configure how this process affects the resulting token including the ability to automatically grant the
audience the client is entitled to request.

[OAuth 2.0 Client Credentials Flow]: #oauth-20-client-credentials-flow

#### OAuth 2.0 Bearer Token Usage

In conjunction with [OAuth 2.0 Client Credentials Flow] special OAuth 2.0 Bearer Access Tokens can be utilized with the
new [OAuth 2.0 Client Credentials Flow] and an additional flow which allows for users to create their own tokens. We
will be adding tooling to be able to do this in the very near future though it's technically already supported via
secure standardized mechanisms. This is implemented as per [RFC6750].

More information on this feature can be found in the [OAuth 2.0 Bearer Token Usage Integration Guide].

[OAuth 2.0 Bearer Token Usage Integration Guide]: ../../integration/openid-connect/oauth-2.0-bearer-token-usage.md
[RFC6750]: https://datatracker.ietf.org/doc/html/rfc6750

#### OAuth 2.0 Authorization Server Issuer Identification

This adds a special response parameter to various flows which identifies the issuer which clients can use to perform
additional verification against as per [RFC9207]. Not all clients support this but those that do are now automatically supported.

[RFC9207]: https://datatracker.ietf.org/doc/html/rfc9207

#### OAuth 2.0 Pushed Authorization Requests

Support for [RFC9126] known as [Pushed Authorization Requests] is one of the main features being added to our
[OpenID Connect 1.0] implementation in this release.

[Pushed Authorization Requests] allows for relying parties / clients to send the Authorization Request parameters over a
back-channel and receive an opaque URI to be used as the `redirect_uri` on the standard Authorization endpoint in place
of the standard Authorization Request parameters.

The endpoint used by this mechanism requires the relying party provides the Token Endpoint authentication parameters.

This means the actual Authorization Request parameters are never sent in the clear over the front-channel. This helps
mitigate a few things:

1. Enhanced privacy. This is the primary focus of this specification.
2. Part of conforming to the [OpenID Connect 1.0] specification [Financial-grade API Security Profile 1.0 (Advanced)].
3. Reduces the attack surface by preventing an attacker from adjusting request parameters prior to the Authorization
   Server receiving them.
4. Reduces the attack surface marginally as less information is available over the front-channel which is the most
   likely location where an attacker would have access to information. While reducing access to information is not
   a reasonable primary security method, when combined with other mechanisms present in [OpenID Connect 1.0] it is
   meaningful.

Even if an attacker gets the [Authorization Code], they are unlikely to have the `client_id` for example, and this is
required to exchange the [Authorization Code] for an [Access Token] and ID Token.

This option can be enforced globally for users who only use relying parties which support
[Pushed Authorization Requests], or can be individually enforced for each relying party which has support.

#### OAuth 2.0 JWT Secured Authorization Response Mode

Also known as JARM, the JWT Secured Authorization Response Mode allows for the entire response from a authorization
server to be formally signed and/or encrypted by a known key. Support for this has been added for clients which support
this mechanism. This can be configured by setting the client configuration option [response_modes] to allow one of the
JARM response modes such as `query.jwt`, `fragment.jwt`, or `form_post.jwt` as well
as setting the client configuration option [authorization_signed_response_alg] or [authorization_signed_response_key_id].

The latter is done via the same means as the [Client JSON Web Key Selection] process.

[response_modes]: ../../configuration/identity-providers/openid-connect/clients.md#response_modes
[authorization_signed_response_alg]: ../../configuration/identity-providers/openid-connect/clients.md#authorization_signed_response_alg
[authorization_signed_response_key_id]: ../../configuration/identity-providers/openid-connect/clients.md#authorization_signed_response_key_id

#### OAuth 2.0 JWT Response for Token Introspection

Similar to the above [OAuth 2.0 JWT Secured Authorization Response Mode](#oauth-20-jwt-secured-authorization-response-mode)
the introspection responses can now be a signed JSON Web Token. This is done similarly via the client configuration options
[introspection_signed_response_alg] and [introspection_signed_response_key_id].

This is done via the same means as the [Client JSON Web Key Selection] process.

[introspection_signed_response_alg]: ../../configuration/identity-providers/openid-connect/clients.md#introspection_signed_response_alg
[introspection_signed_response_key_id]: ../../configuration/identity-providers/openid-connect/clients.md#introspection_signed_response_key_id

#### Proof Key for Code Exchange by OAuth Public Clients

While we already support [RFC7636] commonly known as [Proof Key for Code Exchange], and support enforcement at a global
level for either public clients or all clients, we've added a feature where administrators will be able to enforce
[Proof Key for Code Exchange] on individual clients via the
[require_pkce](../../configuration/identity-providers/openid-connect/clients.md#require_pkce) client configuration option.

It should also be noted that [Proof Key for Code Exchange] can be used at the same time as
[OAuth 2.0 Pushed Authorization Requests](#oauth-20-pushed-authorization-requests).

These features combined with our requirement for the HTTPS scheme are very powerful security measures.

[RFC7636]: https://datatracker.ietf.org/doc/html/rfc7636
[RFC9126]: https://datatracker.ietf.org/doc/html/rfc9126

[Proof Key for Code Exchange]: https://oauth.net/2/pkce/
[Access Token]: https://oauth.net/2/access-tokens/
[Authorization Code]: https://oauth.net/2/grant-types/authorization-code/
[Financial-grade API Security Profile 1.0 (Advanced)]: https://openid.net/specs/openid-financial-api-part-2-1_0.html
[OpenID Connect 1.0]: https://openid.net/
[Pushed Authorization Requests]: https://oauth.net/2/pushed-authorization-requests/

#### Client Authentication Method (Token Endpoint)

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This change may not work for some clients by default as the default option requires
`client_secret_basic` which is the standardized default value most modern clients should use. However, it's really easy
to update. Traditionally this would be considered a breaking change however per our
[Versioning Policy](https://www.authelia.com/policies/versioning/) while we aim to avoid it the OpenID Connect 1.0
implementation is considered excluded at this time.
{{< /callout >}}

This release will allow administrators to configure the Client Authentication Method for the Token Endpoint,
restricting the client usage of the token endpoint and paving the way to more advanced Client Authentication Methods.

This can be configured via the client
[token_endpoint_auth_method](../../configuration/identity-providers/openid-connect/clients.md#token_endpoint_auth_method)
configuration option. The logs should give an indication of the method used by clients so it should be very easy to
update this option.

In addition to this support for new Client Authentication Methods has been added. Specifically, `client_secret_jwt` and
`private_key_jwt`.

#### Subject-Based Client Authorization Policies

Clients can now be configured to allow, disallow, or require a specific authentication level for individual users or
groups of users. See [authorization_policies](../../configuration/identity-providers/openid-connect/provider.md#authorization_policies)
for more information.

#### Per-Client Per-Grant Per-Token Lifespans

Lifespans can now be configured per-client on a per-grant and per-token basis. This allows very fine grained control
over how long a particular token is valid for. For example on client `a`, for the Client Credentials Grant, you can
specifically control the Access Token and Refresh Token lifespans independently. See
[lifespans](../../configuration/identity-providers/openid-connect/provider.md#lifespans) for more information.

#### Additional Client Validations

This release will add additional client configuration validations for various elements which are not technically
compatible. It's important to note that these likely will become errors but are currently just warnings.

#### Multiple JSON Web Keys

The issuer can now be configured with multiple JSON Web Keys which allows signing different clients with different keys
or algorithms depending on specific application requirements or internal policy.

##### Changes {#changes-multiple-json-web-keys}

The following examples illustrate a before and after change for this element in OpenID Connect 1.0. If you do not make
this change many of the new features in OpenID Connect 1.0 that revolve around selecting a key will not be supported
(they may work, but it's not guaranteed, if they're not working you'll be asked to fix this).

The available options and a full example is available from the [jwks] section of the OpenID Connect 1.0 Provider guide.

It should be noted that the new configuration style does not natively support secrets. Instead this must be done via the
template filter example. To use this example you'll need to enable the `template` filter using the
`X_AUTHELIA_CONFIG_FILTERS` environment variable i.e. `X_AUTHELIA_CONFIG_FILTERS=template`. See
[Templating](#templating) for more information.

{{< details "Before" >}}
```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    issuer_private_key: |
      -----BEGIN PRIVATE KEY-----
      ...
      -----END PRIVATE KEY-----
    issuer_certificate_chain: |
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----
```
{{< /details >}}

{{< details "After" >}}
```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    jwks:
      - key: |
          -----BEGIN PRIVATE KEY-----
          ...
          -----END PRIVATE KEY-----
        certificate_chain: |
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
```
{{< /details >}}

{{< details "After (template filter)" >}}
```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    jwks:
      - key: {{ secret "/config/jwks/rsa.2048.pem" | mindent 10 "|" | msquote }}
        certificate_chain: {{ secret "/config/jwks/rsa.2048.cert" | mindent 10 "|" | msquote }}
```
{{< /details >}}

#### Client JSON Web Key Selection

Several client options are now exposed to the administrator allowing configuration of the JSON Web Key used to sign
particular operations, a majority of these can be selected either via the algorithm or the specific key ID providing the
JSON Web Key is registered in the new [jwks] option.

[Client JSON Web Key Selection]: #client-json-web-key-selection
[jwks]: ../../configuration/identity-providers/openid-connect/provider.md#jwks

#### OAuth 2.0 JWT Profile for Access Tokens

Now administrators can configure on a per-client basis the use of the JWT Profile for Access Tokens per [RFC9068]. This
is configured on the registered client level with the [access_token_signed_response_alg] and
[access_token_signed_response_key_id] configuration options.

This is done via the same means as the [Client JSON Web Key Selection] process.

[RFC9068]: https://datatracker.ietf.org/doc/html/rfc9068
[access_token_signed_response_alg]: ../../configuration/identity-providers/openid-connect/clients.md#access_token_signed_response_alg
[access_token_signed_response_key_id]: ../../configuration/identity-providers/openid-connect/clients.md#access_token_signed_response_key_id

#### OAuth 2.0 Authorization Server Metadata and OpenID Connect Discovery 1.0 Signing

The discovery endpoints can now optionally embed a signed JWT into their values for compatible clients to verify the
discovery metadata.

These can be configured via the [discovery_signed_response_alg] and [discovery_signed_response_key_id] configuration
options.

This is done via similar means as the [Client JSON Web Key Selection] process.

[discovery_signed_response_alg]: ../../configuration/identity-providers/openid-connect/provider.md#discovery_signed_response_alg
[discovery_signed_response_key_id]: ../../configuration/identity-providers/openid-connect/provider.md#discovery_signed_response_key_id

#### Other Notable Changes

- The Registered Client `sector_identifier_uri` was previously not validated properly which has been fixed but this may
  require consideration when upgrading. See the
  [sector_identifier_uri](../../configuration/identity-providers/openid-connect/clients.md#sector_identifier_uri)
  configuration document for more information.

## Multi-Domain Protection

In this release we are releasing the main implementation of the Multi-Domain Protection roadmap item.
Please see the [roadmap](../../roadmap/active/multi-domain-protection.md) for more information.

#### Initial Implementation

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This feature at the time of this writing, will not work well with WebAuthn. Steps are being taken
to address this however it will not specifically delay the release of this feature.
{{< /callout >}}

This release sees the initial implementation of multi-domain protection. Users will be able to configure more than a
single root domain for cookies provided none of them are a subdomain of another domain configured. In addition each
domain can have individual settings.

This does not allow single sign-on between these distinct domains. When surveyed users had very low interest in this
feature and technically speaking it's not trivial to implement such a feature as a lot of critical security
considerations need to be addressed.

In addition this feature will allow configuration based detection of the Authelia Portal URI on proxies other than
NGINX/NGINX Proxy Manager/SWAG/HAProxy with the use of the new
[Customizable Authorization Endpoints](#customizable-authorization-endpoints). This is important as it means you only
need to configure a single middleware or helper to perform automatic redirection.

#### Changes {#changes-multiple-domain-protection}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
If you decide to make these changes you should make these at the same time with the
[Updated Proxy Configuration (Customizable Authorization Endpoints)](#changes-customizable-authorization-endpoints).
{{< /callout >}}

The following examples illustrate a before and after change for this element in the session configuration. If you do not
make this change many of the new features in the Forwarded / Redirected Authorization Flow that will never be available.

The available options and a full example is available from the [Session](../../configuration/session/introduction.md)
section of the configuration documentation.

To use the template filter example you'll need to enable the `template` filter using the `X_AUTHELIA_CONFIG_FILTERS`
environment variable i.e. `X_AUTHELIA_CONFIG_FILTERS=template`. See [Templating](#templating) for more information.

{{< details "Before" >}}
```yaml {title="configuration.yml"}
default_redirection_url: 'https://www.example.com'
session:
  name: 'authelia_session'
  domain: 'example.com'
  same_site: 'lax'
  secret: 'insecure_session_secret'
  expiration: '1h'
  inactivity: '5m'
  remember_me_duration: '1M'
```
{{< /details >}}

{{< details "After" >}}
```yaml {title="configuration.yml"}
session:
  secret: 'insecure_session_secret'
  name: 'authelia_session'
  same_site: 'lax'
  inactivity: '5m'
  expiration: '1h'
  remember_me: '1M'
  cookies:
    - domain: 'example.com'
      authelia_url: 'https://auth.example.com'
      default_redirection_url: 'https://www.example.com'
```
{{< /details >}}

{{< details "After (template filter)" >}}
```yaml {title="configuration.yml"}
session:
  secret: 'insecure_session_secret'
  name: 'authelia_session'
  same_site: 'lax'
  inactivity: '5m'
  expiration: '1h'
  remember_me: '1M'
  cookies:
    - domain: '{{ env "DOMAIN_A" }}'
      authelia_url: 'https://auth.{{ env "DOMAIN_A" }}'
      default_redirection_url: 'https://www.{{ env "DOMAIN_A" }}'
```
{{< /details >}}

## WebAuthn

As part of our ongoing effort for comprehensive support for WebAuthn we'll be introducing several important
features. Please see the [roadmap](../../roadmap/complete/webauthn.md) for more information.

#### Multiple WebAuthn Credentials Per-User

In this release we see full support for multiple WebAuthn credentials. This is a fairly basic feature but getting the
frontend experience right is important to us. This is going to be supported via the
[User Control Panel](#user-dashboard--control-panel).

## Customizable Authorization Endpoints

For the longest time we've managed to have the `/api/verify` endpoint perform all authorization verification. This has
served us well however we've been growing out of it. This endpoint is being deprecated in favor of new customizable
per-implementation endpoints. Each existing proxy we support uses one of these distinct implementations.

The old endpoint will still work, in fact you can technically configure an additional endpoint using the methodology of
it via the `Legacy` implementation. However this is strongly discouraged and will not intentionally have new features or
fixes (excluding security fixes) going forward.

In addition to being able to customize them you can create your own, and completely disable support for all other
implementations in the process. Use of these new endpoints will require reconfiguration of your proxy, we plan to
release a guide for each proxy.

See the [Server Authz Endpoints](../../configuration/miscellaneous/server-endpoints-authz.md) guide for more
information.

### Changes {#changes-customizable-authorization-endpoints}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
If you decide to make these changes you should make these at the same time with the
[Session Changes (Multiple Domain Protection)](#changes-multiple-domain-protection)
{{< /callout >}}

It should be noted that making the following changes is strongly recommended to occur at the same time as the
[Multi-Domain Protection](#changes-multiple-domain-protection) changes as several of the features are dependent on the
other.

The main changes which need to occur for everyone is that instead of using the deprecated legacy `/api/verify` endpoint
for their proxy integration they need to upgrade to the `/api/authz/*` variant applicable to their proxy and remove the
`rd` parameter from this integration as this is now handled via the new `authelia_url` value from the session changes.

For example if your previous proxy configuration included `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/verify?rd=https://auth.example.com`
it now by default becomes `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth` for Traefik / Caddy / HAProxy, by default
becomes `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/auth-request` for NGINX based proxies, and by default becomes
`{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/ext-authz` for Envoy.

It should be noted these new endpoints can be customized in the
[server endpoints authz](../../configuration/miscellaneous/server-endpoints-authz.md) section, if custom configuration
is supplied then only the configured endpoints will actually exist, and the proxies may require some additional
configuration depending on your proxy and you should consult the
[Integration Guide](../../integration/proxies/introduction.md) for your particular proxy.


## User Dashboard / Control Panel

As part of our ongoing effort for comprehensive support for a User Dashboard / Control Panel we'll be introducing
several important features. Please see the [roadmap](../../roadmap/active/dashboard-control-panel-for-users.md) for more
information. This has been confused by a few as an Admin Dashboard for configuring settings but this is not the intent.

#### Device Registration OTP

Instead of the current link, in this release users will instead be sent a One Time Code / Password, cryptographically
randomly generated by Authelia. This One Time Password will grant users a duration to perform security sensitive tasks
that we're calling _elevation_.

Naturally how long the code is valid for and the duration the user is considered elevated is customizable in the new
[Elevated Session](../../configuration/identity-validation/elevated-session.md) section.

The motivation for this is that it works in more situations, and is slightly less prone to phishing.

#### TOTP Registration

Instead of just assuming that users have successfully registered their TOTP application, we will require users to enter
the TOTP code prior to it being saved to the database.

## Configuration

Several enhancements are landing for the configuration.

#### Directories

Users will now be able to configure a directory where all `.yml` and `.yaml` files will be loaded in lexical order.
This will not allow combining lists (all of the Access Control Rules must be in the same file, and OpenID Connect 1.0
registered clients must be in the same file), but it will allow you to split portions of the configuration easily.

#### Discovery

Environment variables are being added to assist with configuration discovery, and this will be the default method for
our containers. The advantage is that since the variable will be available when executing commands from the container
context, even if the configuration paths have changed or you've defined additional paths, the `authelia` command will
know where the files are if you properly use this variables.

See the [Loading behavior and Discovery](../../configuration/methods/files.md#loading-behavior-and-discovery) section
of the File Methods guide for more information.

#### Templating

The file based configuration will have access to several experimental templating filters which will assist in creating
configuration templates. The initial one will just expand *most* environment variables into the configuration. The
second will use the go template engine in a very similar way to how Helm operates.

As these features are experimental they may break, be removed, or otherwise not operate as expected. However most of our
testing indicates they're incredibly solid.

See Also:
- [Configuration > Prologue > Security Sensitive Values](../../configuration/prologue/security-sensitive-values.md)
- [Configuration > Methods > Files: File Filters](../../configuration/methods/files.md#file-filters)
- [Reference > Guides > Templating](../../reference/guides/templating.md)

## Miscellaneous

Some miscellaneous notes about this release.

#### Email Notifications

Events triggered by users will generate new notifications sent to their inbox, for example adding a new 2FA device.

#### Storage Import/Export

Utility functions to assist in exporting and subsequently importing the important values in Authelia are being added and
unified in this release.

See the [Authelia CLI Reference Guide](../../reference/cli/authelia/authelia.md) for more information.

#### Privacy Policy

We'll be introducing a feature which allows administrators to more easily comply with the GDPR which optionally shows a
link to their individual privacy policy on the frontend, and optionally requires users to accept it before using
Authelia.

#### LDAP Implementations

This release adds several LDAP implementations into our existing set. See the LDAP configuration option
[implementation](../../configuration/first-factor/ldap.md#implementation) and the
[LDAP Reference Guide](../../integration/ldap/introduction.md) for more information.

#### Server Listener

The server listener configuration was factorized, when updating it to the new values users may be confused about how to
properly do this and when utilizing the old `path` option in bad proxy configurations this may
lead to authorization being skipped. In particular if you're using the path as part of the `/api/verify` endpoint
equivalent.

This doesn't occur if you update the configuration correctly however we have had one user run into this issue. The
following shows how to properly map the old values to he new. For more information see the
[Server Configuration](../../configuration/miscellaneous/server.md#address).

{{< details "Before" >}}
```yaml {title="configuration.yml"}
server:
  host: '0.0.0.0'
  port: {{</* sitevar name="port" nojs="9091" */>}}
  path: 'authelia'
```
{{< /details >}}

{{< details "After" >}}
```yaml {title="configuration.yml"}
server:
  address: 'tcp://0.0.0.0:{{</* sitevar name="port" nojs="9091" */>}}/authelia'
```
{{< /details >}}

