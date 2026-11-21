---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Caddy Security"
description: "A guide on integrating Caddy Security with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Caddy Security | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Caddy Security with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Caddy Security]
  - [v1.6.2](https://github.com/greenpau/caddy-security/releases/tag/v1.6.2)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://caddy-security.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://caddy-security.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2/authelia/authorization-code-callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `caddy-security`
- **Client Secret:** `insecure_secret`
- **Realm:** `authelia`
  - This option determines the redirect URI in the format of
    `https://caddy-security.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2/<realm>/authorization-code-callback`.
    This means if you change this value, you need to update the redirect URI.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Caddy Security] which
will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'caddy-security'
        client_name: 'Caddy Security'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://caddy-security.{{< sitevar name="domain" nojs="example.com" >}}/auth/oauth2/authelia/authorization-code-callback'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="caddy-security" claims="email,name,groups" %}}

### Application

To configure [Caddy Security] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

To configure [Caddy Security] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration
which protects the `/app` path and requires users to be in the `app-members` group:

```caddyfile {title="Caddyfile"}
{
	security {
		oauth identity provider authelia {
			driver generic
			realm authelia
			client_id {env.AUTHELIA_CLIENT_ID}
			client_secret {env.AUTHELIA_CLIENT_SECRET}
			scopes openid email profile groups
			issuer https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
			metadata_url https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration
			login icon text "Authelia"
		}

		authentication portal myportal {
			enable identity provider authelia
			crypto default token lifetime 3600
			crypto key sign-verify {env.JWT_SHARED_KEY}
			cookie path /

			transform user {
				regex match role "^(authp/.*|app/member)$"
				action drop matched role
			}

			transform user {
				match realm authelia
				action add role authp/user
			}

			transform user {
				match realm authelia
				match role app-members
				action add role app/member
			}
		}

		authorization policy apppolicy {
			set auth url https://caddy-security.{{< sitevar name="domain" nojs="example.com" >}}/auth/
			crypto key verify {env.JWT_SHARED_KEY}
			allow roles app/member
		}
	}
}

caddy-security.{{< sitevar name="domain" nojs="example.com" >}} {
	route {
		redir /auth /auth/ 308

		authenticate /auth/* with myportal

		@app path /app /app/*
		route @app {
			authorize with apppolicy
			reverse_proxy app:8080
		}
	}
}
```

The following environment variables must be set for [Caddy]:

```shell {title=".env"}
AUTHELIA_CLIENT_ID=caddy-security
AUTHELIA_CLIENT_SECRET=insecure_secret
JWT_SHARED_KEY=a-long-random-shared-key
```

## See Also

- [Caddy Security Generic OAuth 2.0 / OpenID Connect Documentation](https://docs.authcrunch.com/docs/authenticate/oauth/backend-oauth2-0000-generic)

[Authelia]: https://www.authelia.com
[Caddy]: https://caddyserver.com/
[Caddy Security]: https://github.com/greenpau/caddy-security
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
