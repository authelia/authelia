---
title: "Technitium DNS Server"
description: "Integrating Technitium DNS Server with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2026-06-19T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/technitium/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: ""
  description: ""
  canonical: ""
  noindex: false
---

## Tested Versions

* [Authelia] v4.39.x
* [Technitium DNS Server] v15.2

## Before You Begin

This guide uses the [OpenID Connect 1.0](../../introduction.md) flows. Some [common notes](../introduction.md#common-notes)
apply, in particular regarding generating a [client secret](../../introduction.md#client-secret).

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://dns.example.com/`  (the Technitium web console)
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `technitium`
* __Client Secret:__ `insecure_secret`  (generate your own — see below)
* __Authelia group → Technitium group:__ members of the Authelia group `dns-admins` become Technitium `Administrators`.

## Configuration

### Authelia

The following YAML configures the Authelia [OpenID Connect 1.0 Provider] to register Technitium as a client. Technitium
maps SSO users to local groups from a group claim, so we deliver the user's groups in **both** the standard `groups`
claim and a custom `roles` claim (Technitium's group-claim name is not formally documented; sending both is robust).

```yaml
identity_providers:
  oidc:
    authorization_policies:
      technitium:
        default_policy: 'deny'
        rules:
          - policy: 'two_factor'
            subject:
              - 'group:dns-admins'
    claims_policies:
      technitium:
        id_token:
          - 'groups'
        custom_claims:
          roles:
            attribute: 'groups'
    scopes:
      technitium_roles:
        claims:
          - 'roles'
    clients:
      - client_id: 'technitium'
        client_name: 'Technitium DNS Server'
        client_secret: '$pbkdf2-sha512$310000$...'  # authelia crypto hash generate pbkdf2 --variant sha512
        public: false
        authorization_policy: 'technitium'
        claims_policy: 'technitium'
        require_pkce: false
        redirect_uris:
          - 'https://dns.example.com/sso/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
          - 'technitium_roles'
        grant_types:
          - 'authorization_code'
        response_types:
          - 'code'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure Technitium DNS Server to use Authelia as an OpenID Connect 1.0 Provider:

1. Sign in to the Technitium web console with a **local administrator** account (keep one for break-glass).
2. Go to **Administration** → **Sessions** → **Single Sign-On (SSO)** and enable it.
3. Configure:
   * **Metadata Address:** `https://auth.example.com/.well-known/openid-configuration`
   * **Client ID:** `technitium`
   * **Client Secret:** the secret you generated
   * **Scopes:** `openid`, `profile`, `email`, `groups`, `technitium_roles`
4. **Group Map:** add a mapping — Remote Group `dns-admins` → Local Group `Administrators`.
5. Enable **Allow New User Sign Up** and **Allow Sign Up Only For Mapped Users** (so only members of a mapped group are auto-provisioned).
6. **Save Config** (Technitium restarts its web service automatically).

The redirect URI Technitium uses is `https://<your-console-host>/sso/callback` — it is shown on the SSO settings page; ensure it matches the `redirect_uris` in the Authelia client.

{{< callout context="danger" title="Back-channel name resolution (common pitfall)" icon="outline/alert-triangle" >}}
Technitium DNS Server resolves **its own outbound requests** (the OIDC discovery, token, and JWKS back-channel) using **its own DNS engine**, *not* the operating system's `/etc/hosts` or stub resolver. If your Authelia URL (`auth.example.com`) is **internal-only** (split-horizon / no public record), Technitium will report **"Failed to reach SSO provider"** even though `curl` from the same host succeeds.

**Fix:** make the Authelia hostname resolvable **by Technitium itself** — e.g. add an internal authoritative record for `auth.example.com` on the Technitium server (or via its configured forwarder/conditional path), and confirm with `dig +short @127.0.0.1 auth.example.com` on the Technitium host returning the correct internal IP. Also ensure Technitium can reach the Authelia endpoint on TCP 443.
{{< /callout >}}

## See Also

* [Technitium DNS Server Documentation](https://technitium.com/dns/)
* [Authelia OpenID Connect 1.0 Provider Configuration](../../../../configuration/identity-providers/openid-connect/provider.md)

[Authelia]: https://www.authelia.com
[Technitium DNS Server]: https://technitium.com/dns/
[OpenID Connect 1.0]: ../../introduction.md
[OpenID Connect 1.0 Provider]: ../../../../configuration/identity-providers/openid-connect/provider.md
