---
title: "OAuth 2.0 Bearer Token Usage"
description: "An introduction into utilizing the Authelia OAuth 2.0 Provider as an authorization method"
summary: "An introduction into utilizing the Authelia OAuth 2.0 Provider as an authorization method."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 611
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Access Tokens can be granted which can be leveraged as bearer tokens for the purpose of authorization in place of the
Session Cookie Forwarded Authorization Flow. This is performed leveraging the
[RFC6750: OAuth 2.0 Bearer Token Usage] specification.

## Authorization Endpoints

A [registered OAuth 2.0 client](../../configuration/identity-providers/openid-connect/provider.md#clients) which is
permitted to request the `authelia.bearer.authz` scope can request users grant access to a token which can be used
for the forwarded authentication flow integrated into a proxy (i.e. `access_control` rules) in place of the standard
session cookie-based authorization flow (which redirects unauthorized users) by utilizing
[RFC6750: OAuth 2.0 Bearer Token Usage] authorization scheme norms (i.e. using the bearer scheme).

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
These tokens are not intended for usage with the Authelia API, a separate exclusive scope (or scopes) and
specific audiences will likely be implemented at a later date for this.
{{< /callout >}}

### General Protections

The following protections have been considered:

- There are several safeguards to ensure this Authorization Flow cannot operate accidentally. It must be explicitly
  configured:
  - The authorization endpoint must be explicitly configured to allow the `bearer` scheme. See
    [Authorization Endpoint Configuration](#authorization-endpoint-configuration).
  - Must be utilizing the new session configuration. See [Session Configuration](#session-configuration).
  - The [OpenID Connect 1.0 Provider](../../configuration/identity-providers/openid-connect/provider.md) must be
    configured.
  - One or more [OpenID Connect 1.0 Clients](../../configuration/identity-providers/openid-connect/clients.md) must be
    registered with the `authelia.bearer.authz` scope and relevant required parameters.
  - Additional policy requirements are enforced for the client registrations to ensure as much reasonable protection
    as possible.
- The token must:
  - Be granted the `authelia.bearer.authz` scope.
  - Be presented via the `bearer` scheme in the header matching your server Authorization Endpoints configuration. See
    [Authorization Endpoint Configuration](#authorization-endpoint-configuration).
  - Not be expired, revoked, or otherwise invalid.
  - Actually be an Access Token (tokens with the prefix `authelia_at_`, not tokens with the prefixes `authelia_rt_`
    or `authelia_ac_`).
- Authorizations using this method have special specific processing rules when considering the access control rules:
  - If the token was granted via the `authorization_code` grant then the user who granted the consent for the requested
    scope and audience and their effective authentication level (1FA or 2FA) will be used to match the configured
    access control rules.
  - If the token was granted via the `client_credentials` grant then the token will always be considered as having an
    authentication level of 1FA and when it comes to matching a subject rule a special subject type `oauth2:client:<id>`
    will match the token instead of a user or groups (where `<id>` is the registered client id). See
    [Access Control Configuration](#access-control-configuration).
  - The audience of the token is also considered and if the token does not have an audience which is an exact match or
    the prefix of the URL being requested, the authorization will automatically be denied.
- At this time, each request using this scheme will cause a lookup to be performed on the authentication backend.
- Specific changes to the client registration will result in the authorization being denied, such as:
  - The client is no longer registered.
  - The `authelia.bearer.authz` scope is removed from the registration.
  - The audience which matches the request is removed from the registration.
- The audience of the token must explicitly be requested. Omission of the `audience` parameter may be denied and will
  not grant any audience (thus making it useless) even if the client has been whitelisted for the particular audience.

For example, if `john` consents to grant the token, and it includes the audience `https://app1.{{< sitevar name="domain" nojs="example.com" >}}`, but the
 user `john` is not normally authorized to visit `https://app1.{{< sitevar name="domain" nojs="example.com" >}}` the token will not grant access to this resource.
In addition, if `john` has his access updated via the access control rules, their groups, etc., then this access is
automatically applied to these tokens.

These rules effectively give both administrators and end-users fine-grained control over which endpoints can utilize
this authorization scheme as administrators will be required to allow each individual URL prefix which can be requested
and end users will be able to request individual audiences from the allowed list (effectively narrowing the audience
of the token).

The following recommendations should be considered by users who use this authorization method:

- Using the JWT Profile for Access Tokens effectively makes the introspection stateless and is discouraged for this
  purpose unless you have specific performance issues. We would rather find the cause of the performance issues and
  improve them in an instance where they are noticed.

### Audience Request

While not explicitly part of the specifications, the `audience` parameter can be used during the Authorization Request
phase of the Authorization Code Grant Flow or the Access Token Request phase of the Client Credentials Grant Flow. The
specification leaves it up to Authorization Server policy specifically how audiences are granted, and this seems like a
common practice.

### Authorization Endpoint Configuration

This authorization scheme is not available by default and must be explicitly enabled. The following examples demonstrate
how to enable this scheme (along with the basic scheme). See the
[Server Authz Endpoints](../../configuration/miscellaneous/server-endpoints-authz.md) configuration guide for more
information.

```yaml {title="configuration.yml"}
server:
  endpoints:
    authz:
      forward-auth:
        implementation: 'ForwardAuth'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
              - 'Bearer'
          - name: 'CookieSession'
      ext-authz:
        implementation: 'ExtAuthz'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
              - 'Bearer'
          - name: 'CookieSession'
      auth-request:
        implementation: 'AuthRequest'
        authn_strategies:
          - name: 'HeaderAuthRequestProxyAuthorization'
            schemes:
              - 'Basic'
              - 'Bearer'
          - name: 'CookieSession'
      legacy:
        implementation: 'Legacy'
        authn_strategies:
          - name: 'HeaderLegacy'
          - name: 'CookieSession'
```

### Session Configuration

This feature is only intended to be supported while using the new session configuration syntax. See the example below.

```yaml {title="configuration.yml"}
session:
  secret: 'insecure_session_secret'
  cookies:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      authelia_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      default_redirection_url: 'https://www.{{< sitevar name="domain" nojs="example.com" >}}'
```

### Access Control Configuration

In addition to the restriction of the token audience having to match the target location you must also grant access
in the Access Control section of the configuration either to the user or in the instance of the `client_credentials`
grant the client itself.

It is important to note that the `client_credentials` grant is **always** treated as 1FA, thus only the `one_factor`
policy is useful for this grant type.

```yaml {title="configuration.yml"}
access_control:
  rules:
    ## The 'app1.{{< sitevar name="domain" nojs="example.com" >}}' domain for the user 'john' regardless if they're using OAuth 2.0 or session based flows.
    - domain: 'app1.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      subject: 'user:john'

    ## The 'app2.{{< sitevar name="domain" nojs="example.com" >}}' domain for the 'example-three' client when using the 'client_credentials' grant.
    - domain: 'app2.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
      subject: 'oauth2:client:example-three'
```
### Client Restrictions

In addition to the above protections, this scope **_MUST_** only be configured on clients with strict security rules
which must be explicitly set:

1. Are not configured with any additional scope with the following exceptions:
   - The `offline_access` scope.
2. Have both PAR and PKCE with the `S256` challenge enforced.
3. Have a list of audiences which represent the resources permitted to be allowed by generated tokens.
4. Have the `explicit` consent mode.
5. Only allows the `client_credentials`, or the  `authorization_code` and `refresh_token` grant types.
6. Only allows the `code` response type.
   - This is not relevant for the `client_credentials` grant type.
7. Only allows the `form_post` or `form_post.jwt` response modes.
   - This is not relevant for the `client_credentials` grant type.
8. Must either:
  - Be a public client with the Token Endpoint authentication method `none`. See configuration option
    `token_endpoint_auth_method`.
  - Be a confidential client with the Token Endpoint authentication method `client_secret_basic`, `client_secret_jwt`, or
    `private_key_jwt` configured. See configuration option `token_endpoint_auth_method`.

#### Examples

The following examples illustrate how the [Client Restrictions](#client-restrictions) should be applied to a client.

##### Public Client Example

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'example-one'
        public: true
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'http://localhost/callback'
        scopes:
          - 'offline_access'
          - 'authelia.bearer.authz'
        audience:
          - 'https://app1.{{< sitevar name="domain" nojs="example.com" >}}'
          - 'https://app2.{{< sitevar name="domain" nojs="example.com" >}}'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        response_types:
          - 'code'
        response_modes:
          - 'form_post'
        consent_mode: 'explicit'
        require_pushed_authorization_requests: true
        token_endpoint_auth_method: 'none'
```

##### Confidential Client Example: Authorization Code Flow

This is likely the most common configuration for most users.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'example-two'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'http://localhost/callback'
        scopes:
          - 'offline_access'
          - 'authelia.bearer.authz'
        audience:
          - 'https://app1.{{< sitevar name="domain" nojs="example.com" >}}'
          - 'https://app2.{{< sitevar name="domain" nojs="example.com" >}}'
        grant_types:
          - 'authorization_code'
          - 'refresh_token'
        response_types:
          - 'code'
        response_modes:
          - 'form_post'
        consent_mode: 'explicit'
        require_pushed_authorization_requests: true
        token_endpoint_auth_method: 'client_secret_basic'
```

##### Confidential Client Example: Client Credentials Flow

This example illustrates a method to configure a Client Credential flow for this purpose. This flow is useful for
automations. It's important to note that for access control evaluation purposes this token will match a subject of
`oauth2:client:example-three` i.e. the `oauth2:client:` prefix followed by the client id.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    clients:
      - client_id: 'example-three'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        scopes:
          - 'authelia.bearer.authz'
        audience:
          - 'https://app1.{{< sitevar name="domain" nojs="example.com" >}}'
          - 'https://app2.{{< sitevar name="domain" nojs="example.com" >}}'
        grant_types:
          - 'client_credentials'
        token_endpoint_auth_method: 'client_secret_basic'
```

[RFC6750: OAuth 2.0 Bearer Token Usage]: https://datatracker.ietf.org/doc/html/rfc6750
