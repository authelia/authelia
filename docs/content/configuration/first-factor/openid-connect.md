---
title: "OpenID Connect 1.0"
description: "Configuring an external OpenID Connect 1.0 Provider as a first factor authentication method."
summary: "Authelia supports signing users in with an external OpenID Connect 1.0 Provider. This section describes configuring this."
date: 2026-09-04T09:00:00+10:00
draft: false
images: []
weight: 102400
toc: true
aliases:
  - /c/oidc-rp
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia can act as an [OpenID Connect 1.0] Relying Party, allowing users to sign in with an external
[OpenID Connect 1.0] Provider such as another Authelia instance, an enterprise identity provider, or a public provider.

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
  openid_connect:
    providers:
      - id: 'example'
        name: 'Example'
        issuer: 'https://id.{{< sitevar name="domain" nojs="example.com" >}}'
        client_id: 'authelia'
        client_secret: 'insecure_secret'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        token_endpoint_auth_method: 'client_secret_basic'
        id_token_signed_response_alg: 'RS256'
        pkce:
          challenge_method: 'S256'
        authentication_methods_reference:
          trust: false
        discovery:
          disable: false
        endpoints:
          authorization: ''
          token: ''
          jwks: ''
```

## Options

This section describes the individual configuration options.

### providers

{{< confkey type="list(object)" required="yes" >}}

The list of external providers users may sign in with. At least one provider must be configured when the
`openid_connect` section is present.

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

### name

{{< confkey type="string" required="yes" >}}

The display name for this provider. It is what the sign in button on the login page shows, i.e. a value of `Example`
renders a button labeled `Sign in with Example`.

### issuer

{{< confkey type="string" syntax="uri" required="yes" >}}

The issuer identifier of the external provider. It must be an `https` URI.

This value must be _**exactly**_ the value of the `iss` claim the provider issues, including or omitting a trailing
slash exactly as the provider does. It is compared exactly in three places: against the `issuer` value of the
[discovery](#disable) document, against the `iss` parameter of the callback when the provider sends one, and against
the `iss` claim of the ID Token. It is also the value recorded against every account link.

### client_id

{{< confkey type="string" required="yes" >}}

The client identifier issued to Authelia by the external provider.

### client_secret

{{< confkey type="string" required="situational" >}}

The client secret issued to Authelia by the external provider. This is required unless
[token_endpoint_auth_method](#token_endpoint_auth_method) is `none`.

Unlike the client secrets of the [OpenID Connect 1.0 Provider](../identity-providers/openid-connect/clients.md) role,
this value is the plaintext secret rather than a hash of it, as Authelia must present it to the external provider.

### scopes

{{< confkey type="list(string)" default="openid, profile, email" required="no" >}}

The scopes requested from the external provider. The `openid` scope is mandatory and is automatically prepended to this
list if it is not included in it.

The `profile` and `email` scopes are not required, however the claims they grant are what allows the linking prompt to
show the user which external account is being proposed. Without them the proposal can only show the subject identifier.

### token_endpoint_auth_method

{{< confkey type="string" default="client_secret_basic" required="no" >}}

The client authentication method used at the token endpoint. Must be one of `client_secret_basic`,
`client_secret_post`, or `none`.

### id_token_signed_response_alg

{{< confkey type="string" default="RS256" required="no" >}}

The [JSON Web Signature](https://datatracker.ietf.org/doc/html/rfc7515) algorithm the ID Token must be signed with. An
ID Token signed with any other algorithm is rejected. Must be one of `ES256`, `ES384`, `ES512`, `PS256`, `PS384`,
`PS512`, `RS256`, `RS384`, or `RS512`.

### pkce

Controls [Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636).

#### challenge_method

{{< confkey type="string" default="S256" required="no" >}}

The code challenge method. `S256` is the only permitted value; Authelia always performs
[Proof Key for Code Exchange](https://datatracker.ietf.org/doc/html/rfc7636) and it cannot be disabled.

### authentication_methods_reference

Controls the handling of the `amr` claim from this provider.

#### trust

{{< confkey type="boolean" default="false" required="no" >}}

Trusts the `amr` claim from this provider and merges its values into the session's Authentication Method References.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Authelia does not observe the authentication the user performed at the external provider; it only observes what the
provider asserts about it. Enabling this makes the external provider able to decide the authentication level of the
session, including satisfying a `two_factor` policy. See
[Authentication Level](#authentication-level) for how this interacts with access control.
{{< /callout >}}

### discovery

Controls [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html).

#### disable

{{< confkey type="boolean" default="false" required="no" >}}

Disables discovery. When discovery is enabled, which is the default, Authelia fetches
`<issuer>/.well-known/openid-configuration`, requires that document's `issuer` to exactly match the configured
[issuer](#issuer), and requires it to include the authorization endpoint, the token endpoint, and the JSON Web Key
Set URI.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
Discovery occurs the first time the provider is used rather than during startup, and the result is cached for the
lifetime of the process. An external provider which cannot be reached therefore does not prevent Authelia from
starting or affect any other authentication method; only logins through that provider fail, and the document is
fetched again the next time the provider is used. Failures are recorded in the error level logs.
{{< /callout >}}

When discovery is disabled, [endpoints.authorization](#authorization) and [endpoints.token](#token) are required, and
either [endpoints.jwks](#jwks) or [jwks](#jwks-1) is required.

### endpoints

Explicit endpoints. Any endpoint configured here takes precedence over the equivalent value from discovery, so they may
be used to override an individual endpoint of a provider whose discovery document is otherwise correct. Each endpoint
configured here must be a valid URL with the `https` scheme.

#### authorization

{{< confkey type="string" syntax="uri" required="situational" >}}

The authorization endpoint of the external provider. Required when [discovery.disable](#disable) is `true`.

#### token

{{< confkey type="string" syntax="uri" required="situational" >}}

The token endpoint of the external provider. Required when [discovery.disable](#disable) is `true`.

#### jwks

{{< confkey type="string" syntax="uri" required="situational" >}}

The JSON Web Key Set URI of the external provider, used to fetch the keys which verify the ID Token signature.
Required when [discovery.disable](#disable) is `true` and no inline [jwks](#jwks-1) are configured.

### jwks

{{< confkey type="list(object)" required="no" >}}

Inline JSON Web Keys used to verify the ID Token signature. When any are configured they are used instead of fetching
the key set, and no JSON Web Key Set URI is consulted at all. This is intended for providers which do not publish a key
set over HTTP.

## Redirect URI

The Redirect URI is fixed. It cannot be configured, and it must be registered with the external provider exactly as
below, where `<portal>` is the URL of the Authelia portal and `<id>` is the provider's [id](#id):

```text
https://<portal>/api/firstfactor/openid-connect/<id>/callback
```

For example, a provider with the [id](#id) of `example` on a portal at
`https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` has
the following Redirect URI:

```text
https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/firstfactor/openid-connect/example/callback
```

The Redirect URI is derived from the origin of the request which starts the flow, so a portal reached over more than
one name produces more than one Redirect URI, and each must be registered with the external provider.

## Linking

External identities are anchored to local accounts on the pair of the `iss` and `sub` claims. Neither the `email` claim
nor the `preferred_username` claim participates in finding, matching, or creating a link; they are stored for display
purposes only. This means a change of email address or username at the external provider does not affect an existing
link, and an external account cannot be used to take over a local account which happens to share an email address.

A linking flow can be started from either of two places:

- the provider's button on the login page, which is also how a user with an existing link signs in;
- the _Linked Accounts_ page in the user's settings, which offers a link action for every configured provider the user
  has neither linked nor already has a proposal for.

### Signing in with a link that already exists

1. The user selects the provider's button on the login page.
2. They authenticate at the external provider.
3. The external provider returns them to the [Redirect URI](#redirect-uri).
4. The `iss` and `sub` pair matches an existing link, so they are signed in to that local account and taken to the
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
3. The external provider returns them to the [Redirect URI](#redirect-uri) and the `iss` and `sub` pair matches no
   link.
4. Every value of the validated identity is discarded. The user is returned to the login page with nothing but the
   provider's [id](#id) in the URL, as the `link_provider` query parameter.
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

## Authentication Level

A sign in with an external provider satisfies the `one_factor` policy. It does not satisfy the `two_factor` policy,
regardless of how the user authenticated at the external provider, because Authelia does not observe that
authentication.

To let the external provider's assertion count towards `two_factor`, set
[authentication_methods_reference.trust](#trust) for that provider. The provider's `amr` claim values are then merged
into the session's Authentication Method References, and a claim asserting both a knowledge factor and a possession
factor produces a session which satisfies the `two_factor` policy. This is opt-in per provider because it makes the
external provider trusted to decide the authentication level of the session.

Users may still perform a second factor with Authelia itself after signing in with an external provider, which
satisfies `two_factor` without trusting the provider's assertion at all.

## See Also

- [OpenID Connect 1.0 Provider](../identity-providers/openid-connect/provider.md) for the opposite role, where Authelia
  is the provider other applications rely on.

[OpenID Connect 1.0]: https://openid.net/connect/
