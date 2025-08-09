---
title: "pgAdmin"
description: "Integrating pgAdmin with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-04-13T13:46:05+10:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/pgadmin/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "pgAdmin | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring pgAdmin with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [pgAdmin]
  - [v9.5](https://www.pgadmin.org/docs/pgadmin4/latest/release_notes_9_5.html)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://pgadmin.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `pgadmin`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [pgAdmin] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'pgadmin'
        client_name: 'pgAdmin'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: true
        pkce_challenge_method: 'S256'
        redirect_uris:
          - 'https://pgadmin.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/authorize'
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

### Application

To configure [pgAdmin] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `config_local.py` and in the official container is mounted at `/pgadmin4/`.
{{< /callout >}}

To configure [pgAdmin] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```python {title="config_local.py"}
AUTHENTICATION_SOURCES = ['oauth2', 'internal']
OAUTH2_AUTO_CREATE_USER = True
OAUTH2_CONFIG = [{
	'OAUTH2_NAME': 'Authelia',
	'OAUTH2_DISPLAY_NAME': 'Authelia',
	'OAUTH2_CLIENT_ID': 'pgadmin',
	'OAUTH2_CLIENT_SECRET': 'insecure_secret',
	'OAUTH2_API_BASE_URL': 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}',
	'OAUTH2_AUTHORIZATION_URL': 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization',
	'OAUTH2_TOKEN_URL': 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token',
	'OAUTH2_USERINFO_ENDPOINT': 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo',
	'OAUTH2_SERVER_METADATA_URL': 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration',
	'OAUTH2_SCOPE': 'openid email profile',
	'OAUTH2_USERNAME_CLAIM': 'email',
	'OAUTH2_ICON': 'fa-openid',
	'OAUTH2_BUTTON_COLOR': '<button-color>',
	'OAUTH2_CHALLENGE_METHOD': 'S256',
	'OAUTH2_RESPONSE_TYPE': 'code'
}]
```

## See Also

- [pgAdmin OAuth2 Documentation](https://www.pgadmin.org/docs/pgadmin4/9.5/oauth2.html)

[pgAdmin]: https://www.pgadmin.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
