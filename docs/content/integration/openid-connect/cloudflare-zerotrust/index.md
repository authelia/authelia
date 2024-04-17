---
title: "Cloudflare Zero Trust"
description: "Integrating Cloudflare Zero Trust with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Cloudflare Team Name:__ `example-team`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `cloudflare`
* __Client Secret:__ `insecure_secret`

*__Important Note:__ [Cloudflare Zero Trust] does not properly URL encode the secret per [RFC6749 Appendix B] at the
time this article was last modified (noted at the bottom). This means you'll either have to use only alphanumeric
characters for the secret or URL encode the secret yourself.*

[RFC6749 Appendix B]: https://datatracker.ietf.org/doc/html/rfc6749#appendix-B

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
        redirect_uris:
          - 'https://example-team.cloudflareaccess.com/cdn-cgi/access/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
```

### Application

*__Important Note:__ It is a requirement that the Authelia URL's can be requested by Cloudflare's servers. This usually
means that the URL's are accessible to foreign clients on the internet. There may be a way to configure this without
accessibility to foreign clients on the internet on Cloudflare's end but this is beyond the scope of this document.*

To configure [Cloudflare Zero Trust] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Visit the [Cloudflare Zero Trust Dashboard](https://dash.teams.cloudflare.com)
2. Visit `Settings`
3. Visit `Authentication`
4. Under `Login methods` select `Add new`
5. Select `OpenID Connect`
6. Set the following values:
   1. Name: `Authelia`
   2. App ID: `cloudflare`
   3. Client Secret: `insecure_secret`
   4. Auth URL: `https://auth.example.com/api/oidc/authorization`
   5. Token URL: `https://auth.example.com/api/oidc/token`
   6. Certificate URL: `https://auth.example.com/jwks.json`
   7. Enable `Proof Key for Code Exchange (PKCE)`
   8. Add the following OIDC Claims: `preferred_username`, `mail`
7. Click Save

## See Also

* [Cloudflare Zero Trust Generic OIDC Documentation](https://developers.cloudflare.com/cloudflare-one/identity/idp-integration/generic-oidc/)

[Authelia]: https://www.authelia.com
[Cloudflare]: https://www.cloudflare.com/
[Cloudflare Zero Trust]: https://www.cloudflare.com/products/zero-trust/
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
