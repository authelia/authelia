---
title: "Cloudflare Zero Trust"
description: "Integrating Cloudflare Zero Trust with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/cloudflare-zerotrust/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Cloudflare Zero Trust | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Cloudflare Zero Trust with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.15](https://github.com/authelia/authelia/releases/tag/v4.39.15)

{{% oidc-common bugs="client-credentials-encoding,claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- __Cloudflare Team Name:__ `example-team`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `cloudflare`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Cloudflare] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'cloudflare'
        client_name: 'Cloudflare ZeroTrust'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://example-team.cloudflareaccess.com/cdn-cgi/access/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="cloudflare" %}}

### Application

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
It is a requirement that the Authelia URL's can be requested by Cloudflare's servers. This usually
means that the URLs are accessible to foreign clients on the internet. There may be a way to configure this without
accessibility to foreign clients on the internet on Cloudflare's end, but this is beyond the scope of this document.
{{< /callout >}}

To configure [Cloudflare Zero Trust] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Cloudflare Zero Trust] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Visit the [Cloudflare Zero Trust Dashboard](https://dash.teams.cloudflare.com)
2. Visit `Settings`
3. Visit `Authentication`
4. Under `Login methods` select `Add new`
5. Select `OpenID Connect`
6. Configure the following options:
   - Name: `Authelia`
   - App ID: `cloudflare`
   - Client Secret: `insecure_secret`
   - Auth URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - Certificate URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/jwks.json`
   - Enable `Proof Key for Code Exchange (PKCE)`
   - Add the following OIDC Claims: `preferred_username`, `mail`
7. Click Save

## See Also

- [Cloudflare Zero Trust Generic OIDC Documentation](https://developers.cloudflare.com/cloudflare-one/identity/idp-integration/generic-oidc/)

[Authelia]: https://www.authelia.com
[Cloudflare]: https://www.cloudflare.com/
[Cloudflare Zero Trust]: https://www.cloudflare.com/products/zero-trust/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
