---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "External Identity"
description: "Configuring external identity providers such as OpenID Connect 1.0 Providers, Discord, GitHub, and Plex as a first factor authentication method."
summary: "Authelia supports signing users in with an external identity provider. This section describes configuring this."
date: 2026-09-04T09:00:00+10:00
draft: false
images: []
weight: 102400
toc: true
aliases:
  - /c/oidc-rp
  - /c/external-identity
  - /configuration/first-factor/openid-connect/
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia can sign users in with an external identity provider. The following types of provider are supported:

- `openid_connect`: any [OpenID Connect 1.0] Provider, such as another Authelia instance, an enterprise identity
  provider, or a public provider. Authelia acts as an [OpenID Connect 1.0] Relying Party.
- `discord`: [Discord](https://discord.com/developers/docs/topics/oauth2), which is an OAuth 2.0 provider rather than an
  [OpenID Connect 1.0] Provider.
- `github`: [GitHub](https://docs.github.com/en/apps/oauth-apps), which is an OAuth 2.0 provider rather than an
  [OpenID Connect 1.0] Provider.
- `plex`: [Plex](https://www.plex.tv), which signs users in with their Plex account using the PIN flow Plex apps use
  rather than OAuth 2.0 or [OpenID Connect 1.0].

Step-by-step guides for specific providers are available in the
[External Identity integration](../../integration/external-identity/introduction.md) documentation.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This is _**not**_ a user provider. Authelia never creates users from an external provider. A
[File](file.md) or [LDAP](ldap.md) user provider is still required, and every user who signs in this way must already
exist in it with an account they have linked to their external identity.
{{< /callout >}}

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
authentication_backend:
  external_identity:
    providers:
      - id: 'example'
        type: 'openid_connect'
        name: 'Example'
        issuer: 'https://id.{{< sitevar name="domain" nojs="example.com" >}}'
        client_id: 'authelia'
        client_secret: 'insecure_secret'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_mode: 'query'
        authorization_response_iss_parameter_supported: false
        require_pushed_authorization_requests: false
        token_endpoint_auth_method: 'client_secret_basic'
        id_token_signed_response_alg: 'RS256'
        userinfo_signed_response_alg: ''
        pkce:
          challenge_method: 'S256'
        authentication_methods_reference:
          trust: false
          default: []
          override: false
        discovery:
          disable: false
        endpoints:
          authorization: ''
          token: ''
          userinfo: ''
          jwks: ''
          pushed_authorization_request: ''
      - id: 'discord'
        type: 'discord'
        name: 'Discord'
        client_id: '123456789012345678'
        client_secret: 'insecure_secret'
        scopes:
          - 'identify'
          - 'email'
        token_endpoint_auth_method: 'client_secret_basic'
        authentication_methods_reference:
          default: []
      - id: 'github'
        type: 'github'
        name: 'GitHub'
        client_id: 'Ov23liAbCdEfGhIjKlMn'
        client_secret: 'insecure_secret'
        scopes:
          - 'read:user'
          - 'user:email'
        token_endpoint_auth_method: 'client_secret_basic'
        authentication_methods_reference:
          default: []
      - id: 'plex'
        type: 'plex'
        name: 'Plex'
        client_id: '71f3a2c8-5d4b-4e6a-9c0f-3b8e2d1a7f64'
        authentication_methods_reference:
          default: []
```

## Options

This section describes the individual configuration options. Options which only apply to one type of provider are noted
as such, and configuring an option a type does not support is a configuration error rather than being ignored.

### providers

{{< confkey type="list(object)" required="yes" >}}

The list of external providers users may sign in with. At least one provider must be configured when the
`external_identity` section is present.

### id

{{< confkey type="string" required="yes" >}}

The unique identifier for this provider. It must match the regular expression `^[a-z0-9][a-z0-9_-]{0,31}$`, i.e. it
must begin with a lowercase letter or digit, may otherwise contain lowercase letters, digits, underscores, and hyphens,
and must be at most 32 characters long. It must be unique amongst the configured providers.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This value appears in the [Redirect URI](#redirect-uri) and is recorded against every account link. Changing it after
users have linked their accounts orphans those links and invalidates the [Redirect URI](#redirect-uri) registered with
the external provider.
{{< /callout >}}

### type

{{< confkey type="string" default="openid_connect" required="no" >}}

The type of the external provider. Must be one of `openid_connect`, `discord`, `github`, or `plex`. At most one
`discord`, one `github`, and one `plex` provider may be configured, and every `openid_connect` provider must have a
different [issuer](#issuer), as links are anchored to the issuer.

### name

{{< confkey type="string" required="yes" >}}

The display name for this provider. It is what the sign in button on the login page shows, i.e. a value of `Example`
renders a button labeled `Sign in with Example`.

### client_id

{{< confkey type="string" required="yes" >}}

The client identifier issued to Authelia by the external provider.

For `plex` providers nothing is issued by Plex: this is the client identifier Authelia presents to Plex, which may be
any value such as a UUID. See [Plex](#plex).

### client_secret

{{< confkey type="string" required="situational" >}}

The client secret issued to Authelia by the external provider. This is required unless
[token_endpoint_auth_method](#token_endpoint_auth_method) is `none`.

_This option does not apply to `plex` providers._

Unlike the client secrets of the [OpenID Connect 1.0 Provider](../identity-providers/openid-connect/clients.md) role,
this value is the plaintext secret rather than a hash of it, as Authelia must present it to the external provider.

### scopes

{{< confkey type="list(string)" default="openid, profile, email" required="no" >}}

The scopes requested from the external provider.

- `openid_connect`: the default is `openid`, `profile`, and `email`. The `openid` scope is mandatory and is
  automatically prepended to this list if it is not included in it. The `profile` and `email` scopes are not required,
  however the claims they grant are what allows the linking prompt to show the user which external account is being
  proposed. Without them the proposal can only show the subject identifier.
- `discord`: the default is `identify` and `email`. The `identify` scope is mandatory and is automatically prepended to
  this list if it is not included in it. The `email` scope is not required, and the email address is only shown when
  Discord has verified it.
- `github`: the default is `read:user` and `user:email`. No scope is mandatory, as GitHub grants access to the public
  profile without one. The `user:email` scope is not required, and the email address is only shown when it is the
  primary email address of the account and GitHub has verified it.
- `plex`: scopes do not apply.

### response_mode

{{< confkey type="string" default="query" required="no" >}}

The [response mode](https://openid.net/specs/oauth-v2-multiple-response-types-1_0.html#ResponseModes) the external
provider is asked to deliver the authorization response with. Must be one of `query` or `form_post` for
`openid_connect` providers, and must be `query` for `discord`, `github`, and `plex` providers.

With `query` the provider redirects the browser to the [Redirect URI](#redirect-uri) with the response in the query,
and with `form_post` it has the browser `POST` the response to the same [Redirect URI](#redirect-uri) as a form. An
authorization response delivered any other way than the configured response mode is rejected.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `form_post` response mode is a cross-site `POST` whenever the external provider is not on the same site as
Authelia. Browsers do not send the session cookie with a cross-site `POST` unless the session cookie
[same_site](../session/introduction.md#same_site) option is `none`, and without the session cookie the login
cannot be completed. Use the `query` response mode unless the external provider requires `form_post`.
{{< /callout >}}

### authorization_response_iss_parameter_supported

{{< confkey type="boolean" default="false" required="no" >}}

_This option only applies to `openid_connect` providers._

Declares that the external provider sends the `iss` parameter in the authorization response as described by
[RFC9207](https://datatracker.ietf.org/doc/html/rfc9207), which makes that parameter required. An authorization
response from such a provider which does not include the `iss` parameter is rejected, including error responses.

When [discovery](#disable) is enabled the parameter is also required whenever the discovery document advertises
`authorization_response_iss_parameter_supported` as `true`, regardless of this option. This option is therefore
mostly useful when discovery is disabled, or to require the parameter of a provider which sends it without advertising
that it does.

Regardless of this option, an `iss` parameter which is sent and does not exactly match the configured
[issuer](#issuer) is always rejected.

### require_pushed_authorization_requests

{{< confkey type="boolean" default="false" required="no" >}}

_This option only applies to `openid_connect` providers._

Sends every authorization request as a [Pushed Authorization Request](https://datatracker.ietf.org/doc/html/rfc9126).
The authorization request parameters are sent directly to the provider's Pushed Authorization Request endpoint, which
authenticates Authelia with the configured [token_endpoint_auth_method](#token_endpoint_auth_method), and the browser
is only sent the `client_id` and the `request_uri` the provider issued in return. None of the authorization request
parameters therefore pass through the browser.

The endpoint is the [endpoints.pushed_authorization_request](#pushed_authorization_request) option when it is
configured, and otherwise the `pushed_authorization_request_endpoint` of the [discovery](#disable) document. Pushed
Authorization Requests are also used, regardless of this option, whenever the discovery document advertises
`require_pushed_authorization_requests` as `true`. An advertised endpoint alone does not cause them to be used.

When Pushed Authorization Requests are used and no endpoint is known, the provider fails to resolve.

### token_endpoint_auth_method

{{< confkey type="string" default="client_secret_basic" required="no" >}}

The client authentication method used at the token endpoint. Must be one of `client_secret_basic`,
`client_secret_post`, or `none` for `openid_connect` providers, and one of `client_secret_basic` or
`client_secret_post` for `discord` and `github` providers. It does not apply to `plex` providers.

### pkce

Controls [Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636).

#### challenge_method

{{< confkey type="string" default="S256" required="no" >}}

The code challenge method. `S256` is the only permitted value; Authelia always performs
[Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636) and it cannot be disabled.

### issuer

{{< confkey type="string" syntax="uri" required="situational" >}}

_This option only applies to `openid_connect` providers, for which it is required._

The issuer identifier of the external provider. It must be an `https` URI.

This value must be _**exactly**_ the value of the `iss` claim the provider issues, including or omitting a trailing
slash exactly as the provider does. It is compared exactly in three places: against the `issuer` value of the
[discovery](#disable) document, against the `iss` parameter of the callback when the provider sends one, and against
the `iss` claim of the ID Token. It is also the value recorded against every account link.

### id_token_signed_response_alg

{{< confkey type="string" required="no" >}}

_This option only applies to `openid_connect` providers._

The [JSON Web Signature](https://datatracker.ietf.org/doc/html/rfc7515) algorithm the ID Token must be signed with. An
ID Token signed with any other algorithm is rejected. Must be one of `ES256`, `ES384`, `ES512`, `PS256`, `PS384`,
`PS512`, `RS256`, `RS384`, or `RS512`.

When this option is not configured the algorithm is selected from the `id_token_signing_alg_values_supported` values of
the [discovery](#disable) document. `RS256` is selected when the document includes it or does not include the values at
all, and otherwise the first value Authelia supports is selected in the order the document lists them. The provider
fails to resolve when the document includes none which Authelia supports. When [discovery](#disable) is disabled the
default is `RS256`.

### userinfo_signed_response_alg

{{< confkey type="string" required="no" >}}

_This option only applies to `openid_connect` providers._

The [JSON Web Signature](https://datatracker.ietf.org/doc/html/rfc7515) algorithm the
[UserInfo Response](https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse) must be signed with. Must
be one of `ES256`, `ES384`, `ES512`, `PS256`, `PS384`, `PS512`, `RS256`, `RS384`, or `RS512`. This must match the
`userinfo_signed_response_alg` Authelia is registered with at the provider.

When configured, the UserInfo Response must have the `application/jwt` content type and be signed with this algorithm
by a key from the same JSON Web Key Set the ID Token is verified with. The `iss` and `aud` claims of the response are
optional, however when they are present they must be the configured [issuer](#issuer) and include the
[client_id](#client_id) respectively. When not configured the UserInfo Response must be unsigned JSON. A response in
the other form is rejected in either case.

When [discovery](#disable) is enabled and the discovery document advertises `userinfo_signing_alg_values_supported`,
the configured algorithm must be one of those values.

### authentication_methods_reference

Controls the Authentication Method References a session adopts when a user signs in with this provider. Only the
[default](#default) option applies to `discord`, `github`, and `plex` providers, as none of Discord, GitHub, or Plex
asserts anything about how the user authenticated. These providers always adopt the [default](#default) values, and
when none are configured they adopt `pwd` and `kba`, the values of a password sign in with Authelia.

The values an `openid_connect` provider adopts are decided as follows:

1. When [override](#override) is enabled, the [default](#default) values are adopted.
2. Otherwise, when the provider asserts no `amr` claim, or an empty one, the [default](#default) values are adopted.
3. Otherwise, when [trust](#trust) is enabled, the values of the provider's `amr` claim are adopted.
4. Otherwise, no values are adopted.

#### trust

{{< confkey type="boolean" default="false" required="no" >}}

_This option only applies to `openid_connect` providers._

Trusts the `amr` claim from this provider and merges its values into the session's Authentication Method References.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Authelia does not observe the authentication the user performed at the external provider; it only observes what the
provider asserts about it. Enabling this makes the external provider able to decide the authentication level of the
session, including satisfying a `two_factor` policy. See
[Authentication Level](#authentication-level) for how this interacts with access control.
{{< /callout >}}

#### default

{{< confkey type="list(string)" required="no" >}}

The [RFC8176](https://datatracker.ietf.org/doc/html/rfc8176) Authentication Method Reference values adopted when the
provider asserts no `amr` claim, or in place of the values it asserts when [override](#override) is enabled. The
values must not be empty.

For `discord`, `github`, and `plex` providers these values are always adopted, and when none are configured the
values of a password sign in with Authelia, `pwd` and `kba`, are adopted instead.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
These values are adopted on the administrator's assertion alone, without [trust](#trust) and without Authelia observing
the authentication. Values asserting both a knowledge factor and a possession factor, such as `pwd` and `otp`, produce
a session which satisfies the `two_factor` policy. See [Authentication Level](#authentication-level) for how this
interacts with access control.
{{< /callout >}}

#### override

{{< confkey type="boolean" default="false" required="no" >}}

_This option only applies to `openid_connect` providers._

Adopts the [default](#default) values in place of the values of the provider's `amr` claim, even when the provider
asserts one. When enabled, [default](#default) is required.

### discovery

_This option only applies to `openid_connect` providers._

Controls [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html).

#### disable

{{< confkey type="boolean" default="false" required="no" >}}

Disables discovery. When discovery is enabled, which is the default, Authelia fetches
`<issuer>/.well-known/openid-configuration`, requires that document's `issuer` to exactly match the configured
[issuer](#issuer), and requires it to include the authorization endpoint, the token endpoint, and the JSON Web Key
Set URI.

The configuration is also checked against the capabilities the document advertises. The provider fails to resolve with
an error naming the mismatch when the document advertises `id_token_signing_alg_values_supported` without the
configured [id_token_signed_response_alg](#id_token_signed_response_alg), `code_challenge_methods_supported` without
`S256`, `token_endpoint_auth_methods_supported` without the configured
[token_endpoint_auth_method](#token_endpoint_auth_method), or `response_modes_supported` without `form_post` when that
is the configured [response_mode](#response_mode). A capability the document does not advertise at all is not checked.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Discovery occurs the first time the provider is used rather than during startup, and the result is cached for the
lifetime of the process. An external provider which cannot be reached therefore does not prevent Authelia from
starting or affect any other authentication method; only logins through that provider fail, and the document is
fetched again the next time the provider is used. Failures are recorded in the error level logs.
{{< /callout >}}

When discovery is disabled, [endpoints.authorization](#authorization) and [endpoints.token](#token) are required, and
either [endpoints.jwks](#jwks) or [jwks](#jwks-1) is required.

### endpoints

_This option only applies to `openid_connect` providers._

Explicit endpoints. Any endpoint configured here takes precedence over the equivalent value from discovery, so they may
be used to override an individual endpoint of a provider whose discovery document is otherwise correct. Each endpoint
configured here must be a valid URL with the `https` scheme.

#### authorization

{{< confkey type="string" syntax="uri" required="situational" >}}

The authorization endpoint of the external provider. Required when [discovery.disable](#disable) is `true`.

#### token

{{< confkey type="string" syntax="uri" required="situational" >}}

The token endpoint of the external provider. Required when [discovery.disable](#disable) is `true`.

#### userinfo

{{< confkey type="string" syntax="uri" required="no" >}}

The UserInfo endpoint of the external provider. When discovery is enabled this defaults to the `userinfo_endpoint` of
the discovery document.

When the provider has a UserInfo endpoint, Authelia requests it with the access token once the ID Token has been
validated. The `sub` claim of the response must exactly match the `sub` claim of the ID Token or the login is
rejected. The `preferred_username`, `name`, and `email` claims of the response are used in place of those of the ID
Token, and the claims of the ID Token are only used where the response does not include them. When the provider has
no UserInfo endpoint the claims are taken from the ID Token alone.

#### jwks

{{< confkey type="string" syntax="uri" required="situational" >}}

The JSON Web Key Set URI of the external provider, used to fetch the keys which verify the ID Token signature.
Required when [discovery.disable](#disable) is `true` and no inline [jwks](#jwks-1) are configured.

#### pushed_authorization_request

{{< confkey type="string" syntax="uri" required="situational" >}}

The Pushed Authorization Request endpoint of the external provider. When discovery is enabled this defaults to the
`pushed_authorization_request_endpoint` of the discovery document. Required when
[require_pushed_authorization_requests](#require_pushed_authorization_requests) is `true` and
[discovery.disable](#disable) is `true`.

### jwks

{{< confkey type="list(object)" required="no" >}}

_This option only applies to `openid_connect` providers._

Inline JSON Web Keys used to verify the ID Token signature. When any are configured they are used instead of fetching
the key set, and no JSON Web Key Set URI is consulted at all. This is intended for providers which do not publish a key
set over HTTP.

## Discord

A `discord` provider signs users in with [Discord](https://discord.com/developers/docs/topics/oauth2) using the OAuth
2.0 authorization code flow with [Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636). Discord's
endpoints are fixed, so none of the endpoint or discovery options apply.

To configure it, create an application in the
[Discord Developer Portal](https://discord.com/developers/applications), add the [Redirect URI](#redirect-uri) to its
OAuth2 redirects, and use the application's client ID and client secret as the [client_id](#client_id) and
[client_secret](#client_secret). See the [Discord integration guide](../../integration/external-identity/discord.md)
for step-by-step instructions.

Discord is not an [OpenID Connect 1.0] Provider, so there is no ID Token to validate. The identity is instead the
Discord user the access token belongs to, as returned by Discord's current user endpoint, with the access token
obtained directly from Discord's token endpoint over TLS. Accounts are linked on the Discord user ID, which never
changes; the username, display name, and email address are only displayed.

## GitHub

A `github` provider signs users in with [GitHub](https://docs.github.com/en/apps/oauth-apps) using the OAuth 2.0
authorization code flow with [Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636). GitHub's
endpoints are fixed, so none of the endpoint or discovery options apply.

To configure it, register an OAuth App or a GitHub App with GitHub, use the [Redirect URI](#redirect-uri) as its
callback URL, and use its client ID and client secret as the [client_id](#client_id) and
[client_secret](#client_secret). See the [GitHub integration guide](../../integration/external-identity/github.md) for
step-by-step instructions.

GitHub is not an [OpenID Connect 1.0] Provider, so there is no ID Token to validate. The identity is instead the GitHub
user the access token belongs to, as returned by GitHub's authenticated user endpoint, with the access token obtained
directly from GitHub's token endpoint over TLS. Accounts are linked on the GitHub user ID, which never changes; the
username, display name, and email address are only displayed. The email address is the primary email address of the
account from GitHub's user emails endpoint, and is only used when GitHub has verified it.

## Plex

A `plex` provider signs users in with their [Plex](https://www.plex.tv) account. Plex does not offer OAuth 2.0 or
[OpenID Connect 1.0] to third parties, so users are signed in with the PIN flow Plex apps use instead:

1. Authelia creates a PIN with Plex.
2. The user is sent to Plex to sign in and approve the PIN, and Plex returns them to the [Redirect URI](#redirect-uri).
3. Authelia retrieves the token Plex issued for the approved PIN directly from Plex over TLS, and uses it to retrieve
   the Plex account it belongs to.

There is nothing to register with Plex and no client secret. The [client_id](#client_id) is the client identifier
Authelia presents to Plex, which Plex shows the user as a device named `Authelia`. It may be any value, such as a UUID,
but it should be unique to this Authelia instance and should not change, as Plex treats every value as a different
device. None of the scope, token endpoint, PKCE, endpoint, or discovery options apply. See the
[Plex integration guide](../../integration/external-identity/plex.md) for step-by-step instructions.

The PIN is retained in the session of the user who started the flow, and the PIN the browser returns with must be that
PIN, so a PIN approved in any other flow is rejected. Accounts are linked on the Plex account UUID, which never changes;
the username, display name, and email address are only displayed.

Authelia describes itself to Plex with the Authelia version, a device name of `Authelia (<origin>)` where `<origin>` is
the origin of the portal the flow was started from, and the language the user chose in the portal. Plex shows these
when the user approves the PIN and in the list of devices authorized to their Plex account.

## Trusted Certificates

Every request Authelia makes to the external provider is made over TLS and verifies the certificate of the provider.
This includes the discovery document, the JSON Web Key Set, the token endpoint, the UserInfo endpoint, Discord's current
user endpoint, GitHub's user and user emails endpoints, and Plex's PIN and user endpoints. The certificate is trusted
when it is issued by an authority trusted by the system, or by a certificate in the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory). A provider whose certificate is
issued by a private certificate authority is therefore trusted by placing that certificate authority in the
[certificates_directory](../miscellaneous/introduction.md#certificates_directory); certificate verification cannot be
disabled.

## Redirect URI

The Redirect URI is fixed. It cannot be configured, and it must be registered with the external provider exactly as
below, where `<portal>` is the URL of the Authelia portal including any path it is served under, and `<id>` is the
provider's [id](#id):

```text
https://<portal>/api/firstfactor/external-identity/<id>/callback
```

For example, a provider with the [id](#id) of `example` on a portal at
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` has
the following Redirect URI:

```text
https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/firstfactor/external-identity/example/callback
```

The Redirect URI is derived from the origin of the request which starts the flow, so a portal reached over more than
one name produces more than one Redirect URI, and each must be registered with the external provider.

## Linking

External identities are anchored to local accounts on the [type](#type) of the provider together with the issuer and
the subject. The following table lists the attribute each type of provider takes each value from. Only the issuer and
the subject identify the account; the username, display name, and email are only displayed.

| Type             | Issuer                | Subject                   | Username                       | Display Name                                 | Email                                                                       |
| :--------------- | :-------------------- | :------------------------ | :----------------------------- | :------------------------------------------- | :-------------------------------------------------------------------------- |
| `openid_connect` | the `iss` claim       | the `sub` claim           | the `preferred_username` claim | the `name` claim                             | the `email` claim                                                           |
| `discord`        | `https://discord.com` | the `id` of the user      | the `username` of the user     | the `global_name`, or `username` when absent | the `email` of the user, only when `verified` is `true`                     |
| `github`         | `https://github.com`  | the `id` of the user      | the `login` of the user        | the `name`, or `login` when absent           | the `email` of the user emails entry which is both `primary` and `verified` |
| `plex`           | `https://plex.tv`     | the `uuid` of the account | the `username` of the account  | the `title`, or `username` when absent       | the `email` of the account                                                  |

The username, display name, and email of an `openid_connect` provider are the claims of the UserInfo response, and
are only taken from the ID Token when the UserInfo response does not include them or the provider has no UserInfo
endpoint. The issuer and subject are the claims of the ID Token. The attributes of a `discord` provider are
those of the user returned by Discord's current user endpoint, of a `github` provider those of the user returned by
GitHub's authenticated user endpoint and its user emails endpoint, and of a `plex` provider those of the account
returned by Plex's user endpoint. The GitHub `id` is a number, and is recorded as its decimal representation.

A link is only ever matched when all three are the same, so an identity from a provider of one type never signs in with
a link made through a provider of another type, even if the issuer and subject are the same. All three are covered by
the signature every link is stored with.

Neither the email address nor the username participates in finding, matching, or creating a link; they are stored for
display purposes only. This means a change of email address or username at the external provider does not affect an
existing link, and an external account cannot be used to take over a local account which happens to share an email
address. The email address is recorded whenever the provider returns one, and is otherwise left blank: an
`openid_connect` provider only returns it when the `email` scope is requested and granted, a `discord` provider only
when Discord has verified it, a `github` provider only when the `user:email` scope is granted and GitHub has verified
the primary email address, and a `plex` provider always returns it.

A linking flow can be started from either of two places:

- the provider's button on the login page, which is also how a user with an existing link signs in;
- the _Linked Accounts_ page in the user's settings, which offers a link action for every configured provider the user
  has neither linked nor already has a proposal for.

### Signing in with a link that already exists

1. The user selects the provider's button on the login page.
2. They authenticate at the external provider.
3. The external provider returns them to the [Redirect URI](#redirect-uri).
4. The issuer and subject pair matches an existing link, so they are signed in to that local account and taken to the
   resource they were trying to reach.

### Linking while signed in

1. From the _Linked Accounts_ page the user selects the link action for the provider.
2. They authenticate at the external provider.
3. The external provider returns them to the [Redirect URI](#redirect-uri), and the validated identity is held as a
   proposal against the session they are signed in to.
4. They are returned to the _Linked Accounts_ page where the proposal is shown, and they accept or decline it.

### Linking from the login page

The same flow started from the login page reaches the [Redirect URI](#redirect-uri) with no one signed in, and the
proposal in step 3 above has no account to be held against. Rather than carry the identity across the sign in which
must follow, Authelia discards it and has the user perform the flow again once they are signed in:

1. The user selects the provider's button on the login page.
2. They authenticate at the external provider.
3. The external provider returns them to the [Redirect URI](#redirect-uri) and the issuer and subject pair matches no
   link.
4. Every value of the validated identity is discarded. The user is sent to a dedicated link page, which carries
   nothing but the provider's [id](#id) in the URL as the `link_provider` query parameter. The page explains that the
   external account is not linked yet, that they must sign in to link it, and that they will return to the external
   provider to confirm. The external sign in buttons are not shown on this page, and a link returns to the normal login
   page.
5. They sign in with their Authelia account.
6. Once the authentication their policy requires is complete, the portal performs the flow again on their behalf. The
   external provider usually does not prompt a second time, because the session they established there in step 2 still
   exists, so this is not normally visible to them.
7. The proposal, now produced by a flow the signed in user performed, is shown on the _Linked Accounts_ page for them
   to accept or decline.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
A user who still owes a second factor after step 5 is not interrupted: the flow is only performed again once the
authentication their policy requires is complete.
{{< /callout >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
No value describing the external identity crosses the sign in boundary. Only the provider's [id](#id) does, and it is
taken from the configured provider rather than from the request. This is what prevents someone who can plant a session
cookie on the Authelia domain from planting a proposal for an external account they control, which a user might then
accept believing it to be their own.
{{< /callout >}}

Accepting a proposal and removing an existing link both require an elevated session, i.e. the user must confirm their
identity again before either takes effect. Declining a proposal requires no elevation since it can only remove a
proposal.

Users manage their links from the _Linked Accounts_ page in their settings, which lists each linked provider along with
when the link was created and when it was last used to sign in. Only one link per provider per user is permitted, so a
provider which is already linked is not offered again.

When a sign in with an external provider is rejected for any reason, the user is returned to the login page with a
generic error notification. The reason is only recorded in the logs.

## Authentication Level

A sign in with an external provider satisfies the `one_factor` policy. It does not satisfy the `two_factor` policy,
regardless of how the user authenticated at the external provider, because Authelia does not observe that
authentication.

A sign in with a `discord`, `github`, or `plex` provider is treated as a password sign in with Authelia unless
[authentication_methods_reference.default](#default) is configured: the session adopts `pwd` and `kba`, exactly as a
username and password sign in does. A resource with the `two_factor` policy therefore asks the user for a second factor
with Authelia, just as it does after a password sign in.

To let an `openid_connect` provider's assertion count towards `two_factor`, set
[authentication_methods_reference.trust](#trust) for that provider. The provider's `amr` claim values are then merged
into the session's Authentication Method References, and a claim asserting both a knowledge factor and a possession
factor produces a session which satisfies the `two_factor` policy. This is opt-in per provider because it makes the
external provider trusted to decide the authentication level of the session. None of Discord, GitHub, or Plex asserts
anything about how the user authenticated, so a `discord`, `github`, or `plex` provider never counts towards
`two_factor` on its own assertion.

The [authentication_methods_reference.default](#default) values count towards `two_factor` in the same way, for any
type of provider, whenever they are adopted. This lets an administrator who knows how a provider authenticates its users
decide the authentication level of the sessions it produces, including for providers which assert no `amr` claim.

Users may still perform a second factor with Authelia itself after signing in with an external provider, which
satisfies `two_factor` without trusting the provider's assertion at all.

## See Also

- [External Identity integration](../../integration/external-identity/introduction.md) for step-by-step guides for
  specific providers.
- [OpenID Connect 1.0 Provider](../identity-providers/openid-connect/provider.md) for the opposite role, where Authelia
  is the provider other applications rely on.

[OpenID Connect 1.0]: https://openid.net/connect/
