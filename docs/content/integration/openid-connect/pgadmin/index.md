---
title: "pgAdmin"
description: "Integrating pgAdmin with the Authelia OpenID Connect 1.0 Provider."
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
* [pgAdmin]
  * [v8.5](https://www.pgadmin.org/docs/pgadmin4/8.5/index.html)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://pgadmin.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `pgadmin`
* __Client Secret:__ `insecure_secret`

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
        redirect_uris:
          - 'https://pgadmin.example.com/oauth2/authorize'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [pgAdmin] to utilize Authelia as an [OpenID Connect 1.0] Provider:

1. Add the following YAML to your configuration:

```python {title="config_local.py"}
AUTHENTICATION_SOURCES = ['oauth2', 'internal']
OAUTH2_AUTO_CREATE_USER = True
OAUTH2_CONFIG = [{
	'OAUTH2_NAME': 'Authelia',
	'OAUTH2_DISPLAY_NAME': 'Authelia',
	'OAUTH2_CLIENT_ID': 'pgadmin',
	'OAUTH2_CLIENT_SECRET': 'insecure_secret',
	'OAUTH2_API_BASE_URL': 'https://auth.example.com',
	'OAUTH2_AUTHORIZATION_URL': 'https://auth.example.com/api/oidc/authorization',
	'OAUTH2_TOKEN_URL': 'https://auth.example.com/api/oidc/token',
	'OAUTH2_USERINFO_ENDPOINT': 'https://auth.example.com/api/oidc/userinfo',
	'OAUTH2_SERVER_METADATA_URL': 'https://auth.example.com/.well-known/openid-configuration',
	'OAUTH2_SCOPE': 'openid email profile',
	'OAUTH2_USERNAME_CLAIM': 'email',
	'OAUTH2_ICON': 'fa-key',
	'OAUTH2_BUTTON_COLOR': '<button-color>'
}]
```

## See Also

- [pgAdmin OAuth2 Documentation](https://www.pgadmin.org/docs/pgadmin4/8.4/oauth2.html)

[pgAdmin]: https://www.pgadmin.org/
[Authelia]: https://www.authelia.com
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
