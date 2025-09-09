---
title: "Cloud Identity Engine"
description: "Integrating Cloud Identity Engine with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Cloud Identity Engine | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Cloud Identity Engine with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [Cloud Identity Engine]

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `cloudidentityengine`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Cloud Identity Engine] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'cloudidentityengine'
        client_name: 'Cloud Identity Engine'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - '' # Replace with the value copied in step 7.
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Cloud Identity Engine] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Cloud Identity Engine] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to your Cloud Identity Engine administrator account.
2. Select `Authentication`.
3. Select `Authentication Types`.
4. Select `Add New Authentication Type`.
5. Select `Set Up` under `OIDC`.
6. Enter the following values:
   - Authentication Type Name: `Authelia`
   - Client Name: `Authelia`
   - Client ID: `cloudidentityengine`
   - Client Secret: `insecure_secret`
   - OIDC Issuer URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - JWT Encryption Algorithm: `RS256`
   - OIDC Authentication Server Discovery Endpoint (Optional): `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`.
7. Click the copy button in the `Callback URL / Redirect URL` and use this in your `redirect_uris` Authelia configuration.
8. Click `Submit`.

## See Also

- [Cloud Identity Engine Configure Unlock Cloud Identity Engine with SSO using OpenID Connect Documentation](https://docs.paloaltonetworks.com/cloud-identity/cloud-identity-engine-getting-started/authenticate-users-with-the-cloud-identity-engine/set-up-oidc-authentication)

[Authelia]: https://www.authelia.com
[Cloud Identity Engine]: https://docs.paloaltonetworks.com/pan-os/10-1/pan-os-new-features/identity-features/cloud-identity-engine
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
